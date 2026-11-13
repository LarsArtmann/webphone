// Package app is webphone's composition root: one samber/do v2
// container wiring every service, the go-health probe over the
// container's health-checkable services, and the optional
// go-health-dashboard operator surface. cmd/webphone owns process
// concerns (config, logging, signals, the HTTP listener); this package
// owns object lifetime.
//
// Design rules (samber/do best practices, DO-1..DO-6):
//   - the injector lives ONLY here — services hold resolved
//     dependencies, never the container (no service-locator smell);
//   - every do.New() has exactly one Shutdown(), reachable from
//     App.Shutdown on every exit path (DO-2);
//   - Override* appears nowhere in production code (DO-3);
//   - MustInvoke appears only inside provider closures and New
//     itself — the composition-root equivalents of main (DO-1);
//   - shutdowns are self-contained: Database.Shutdown closes the
//     handle, Dashboard.Shutdown drains the pusher (DO-6).
//
// The service packages stay framework-free: they implement
// HealthCheck/Shutdown as plain methods, and the structural
// conformance to samber/do's lifecycle interfaces is asserted HERE,
// at the adapter.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
	"github.com/samber/do/v2"

	"github.com/larsartmann/webphone/internal/admin"
	"github.com/larsartmann/webphone/internal/asr"
	"github.com/larsartmann/webphone/internal/authmap"
	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/paperless"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/push"
	"github.com/larsartmann/webphone/internal/readmodels"
	"github.com/larsartmann/webphone/internal/retention"
	"github.com/larsartmann/webphone/internal/server"
	serverauth "github.com/larsartmann/webphone/internal/server/auth"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/sso"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/larsartmann/webphone/internal/textagent"
	"github.com/larsartmann/webphone/internal/userauth"
	"github.com/larsartmann/webphone/internal/web/views"
)

// Adapter-side lifecycle guards (the service packages stay free of
// framework imports; the container contract is asserted where the
// container lives). store.Database additionally carries the shutdown
// cascade for the SQLite handle.
var (
	_ do.HealthcheckerWithContext = (*store.Database)(nil)
	_ do.HealthcheckerWithContext = (*blob.Store)(nil)
	_ do.HealthcheckerWithContext = (*userauth.Service)(nil)
	_ do.HealthcheckerWithContext = (*crm.Resolver)(nil)
	_ do.HealthcheckerWithContext = (*CqrsService)(nil)
	_ do.ShutdownerWithError      = (*store.Database)(nil)
	_ do.ShutdownerWithError      = (*userauth.Service)(nil)
	_ do.ShutdownerWithError      = (*CqrsService)(nil)
	_ do.Shutdowner               = (*dashboard.Dashboard)(nil)
	_ do.Shutdowner               = (*server.ExtensionHubs)(nil)
)

// probeRefresh runs the probe's background cache loop while the
// dashboard is enabled: its pusher reads CachedResponse on every tick,
// which in live mode (interval 0) would forever hold the boot snapshot.
// Two local checks per second are noise. Disabled dashboard → interval
// 0, exactly the pre-container live-mode behavior.
const probeRefresh = time.Second

// userauthBootTimeout bounds the usermgmt startup drain (projection
// workers replaying the event journal into the SQL read models). The
// journal of a phone-system user list is tiny; the bound exists so a
// broken store fails the boot instead of hanging it.
const userauthBootTimeout = 30 * time.Second

// trendSamples bounds the dashboard's status-history ring (the trend
// card + /health/trend export). One sample per probe tick: a bit over
// four minutes of history at the default cadence.
const trendSamples = 240

// App is the wired application: the container plus the few handles the
// process layer needs (the HTTP handler to serve, the probe to start).
// Build it with New, run background loops with Start, and ALWAYS call
// Shutdown exactly once after serving ends.
type App struct {
	injector *do.RootScope
	cfg      config.Config
	log      *slog.Logger
	// Probe serves /livez and /startupz (handed to server.Deps). It is
	// deliberately NOT registered in the injector it reads: the probe
	// iterates every container service implementing the health-check
	// interface, and *health.Probe itself conforms — self-registration
	// would make every batch recurse into itself.
	Probe *health.Probe
	// Trace is the OPTIONAL flight recorder (nil — the default). It is
	// built outside the injector because the injector's Shutdown would
	// call Close on a typed nil; the process layer owns its lifetime.
	Trace   *traceRecorder
	Handler http.Handler
	// Hubs is the per-extension SSE hub registry, exposed so the process
	// layer can close live streams BEFORE the HTTP drain (ServeSSE
	// returns the moment its subscriber channel closes — an open tab
	// cannot hold the drain for the full shutdown budget). The container
	// cascade closes it again idempotently; the field exists because the
	// drain runs BEFORE do.Shutdown.
	Hubs *server.ExtensionHubs
	// Push is the OPTIONAL push fanout (nil-safe by the seam's own
	// doctrine). The field exists so Start can arm the prune registry
	// once the cqrs system — which the push-devices store rides — has
	// resolved; the HTTP listener accepts only afterwards, so every Fan
	// sees an armed registry.
	Push *push.Fanout
}

// pushRegistryAdapter adapts the journal-backed push-devices store to
// the fanout's prune seam: the fanout stays dependency-light (domain
// types only), and the conversion is the adapter's whole job.
type pushRegistryAdapter struct {
	devices readmodels.PushDeviceGate
}

func (a pushRegistryAdapter) List(ctx context.Context, owner domain.Extension) ([]push.Device, error) {
	views, err := a.devices.List(ctx, owner)
	if err != nil {
		return nil, err
	}
	out := make([]push.Device, 0, len(views))
	for _, v := range views {
		// The row's transport string passed the registration door's
		// ParseTransport at upsert; a row that predates the door (or a
		// hand-seeded test row) parses or skips — never a fan-time
		// lookup miss.
		transport, err := push.ParseTransport(v.Transport)
		if err != nil {
			continue
		}
		out = append(out, push.Device{DeviceID: v.DeviceID, Transport: transport, Token: v.Token})
	}
	return out, nil
}

func (a pushRegistryAdapter) Remove(ctx context.Context, owner domain.Extension, deviceID string) error {
	return a.devices.Remove(ctx, owner, deviceID)
}

// wrapf is this package's home for family-neutral startup wraps (same
// contract as cmd/webphone's propagatef): each step wraps an inner
// error whose family varies by cause, so a fixed-family wrap here would
// clobber the constructor's classification.
func wrapf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// New builds the container: registers every service, eagerly
// instantiates the critical pair and the HTTP handler (fail fast, with
// named errors — lazy leftovers would lie to the startup probe), and
// mounts the dashboard when cfg.Dashboard.Enable is set.
func New(cfg config.Config, log *slog.Logger) (*App, error) {
	// A missing data dir would otherwise surface as SQLite's cryptic
	// "unable to open database file (14)".
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return nil, wrapf("create data dir: %w", err)
	}

	// Wall-clock zone (plan T26d): the configured IANA zone owns every
	// rendered time and log line (validation already rejected typos).
	if cfg.Timezone != "" {
		loc, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			return nil, wrapf("load timezone: %w", err)
		}
		time.Local = loc
		log.Info("timezone applied", "zone", cfg.Timezone)
	}

	injector := do.New()

	// --- infrastructure ------------------------------------------------
	// The critical pair carries the check names the probe classifies by
	// (health.WithCriticalServices below) — the names are contract. Both
	// members ARE constructed during New (the service providers and the
	// handler below MustInvokeNamed them), so HW-4's lazy-registration
	// warning cannot apply — the linter cannot see transitive resolution.
	//samber-linter:allow hw-4 eagerly resolved via MustInvokeNamed in New
	do.ProvideNamed(injector, "sqlite", func(i do.Injector) (*store.Database, error) {
		return store.OpenDatabase(filepath.Join(cfg.DataDir, "webphone.db"))
	})
	//samber-linter:allow hw-4 eagerly resolved via MustInvokeNamed in New
	do.ProvideNamed(injector, "blob-dir", func(i do.Injector) (*blob.Store, error) {
		return blob.New(filepath.Join(cfg.DataDir, "files"))
	})
	// The system composition root (SUPERB plan): the command/event
	// journal rides its OWN system.db beside the read-model engine —
	// engine DDL never meets the hand ladder, and the container sees
	// one more named eager service with probe-discoverable health
	// (optional non-critical leg while only advisory features ride it).
	registerCqrs(injector, cfg, log)
	do.Provide(injector, func(i do.Injector) (*pbx.Client, error) {
		return pbx.NewClient(cfg.PhoneAPIURL)
	})
	//samber-linter:allow hw-4 resolved eagerly via the named "crm" alias below when the seam is on; off = no crm leg anywhere
	do.Provide(injector, func(i do.Injector) (*crm.Client, error) {
		return crm.NewClient(cfg.CRM.URL, cfg.CRM.Token, cfg.CRM.CAFile)
	})
	//samber-linter:allow hw-4 resolved eagerly via the named "crm" alias below when the seam is on (probe discovery)
	do.Provide(injector, func(i do.Injector) (*crm.Resolver, error) {
		return crm.NewResolver(do.MustInvoke[*crm.Client](i), cfg.CRM.ContactURL, log), nil
	})
	// Optional speech-to-text seam (live transcription of every audio
	// stream). Config-absent → a disabled provider (Enabled() false):
	// /api/transcribe 404s and the island hides its affordances. The
	// provider kind is a config decision: the OpenAI-compatible wire
	// (Speaches, whisper.cpp, …) or Google Cloud Speech-to-Text V2. The
	// SAME instance is additionally registered NAMED "asr" when the
	// seam is on: the go-health probe then discovers its
	// outcomes-based HealthCheck as an OPTIONAL NON-CRITICAL leg
	// (#22) — never probed, never gating livez/startupz, visible on the
	// dashboard.
	asrProvider, err := func() (asr.Provider, error) {
		if cfg.ASR.Google() {
			return asr.NewGoogleClient(asr.GoogleConfig{
				Endpoint: cfg.ASR.URL,
				APIKey:   cfg.ASR.Token,
				Project:  cfg.ASR.Project,
				Location: cfg.ASR.Location,
				Model:    cfg.ASR.Model,
				Language: cfg.ASR.Language,
			})
		}
		if cfg.ASR.Provider == "whispercpp" {
			return asr.NewWhisperCppClient(asr.ClientConfig{BaseURL: cfg.ASR.URL, Token: cfg.ASR.Token, Model: cfg.ASR.Model})
		}
		return asr.NewClient(asr.ClientConfig{BaseURL: cfg.ASR.URL, Token: cfg.ASR.Token, Model: cfg.ASR.Model})
	}()
	if err != nil {
		return nil, wrapf("open asr provider: %w", err)
	}
	do.Provide(injector, func(i do.Injector) (asr.Provider, error) { return asrProvider, nil })
	if cfg.ASR.Enabled() {
		//samber-linter:allow hw-4 eagerly invoked below for probe discovery
		do.ProvideNamed(injector, "asr", func(i do.Injector) (asr.Provider, error) { return asrProvider, nil })
	}
	if cfg.CRM.URL != "" {
		// The CRM resolver joins the go-health probe discovery the same
		// way asr does: NAMED registration only when the seam is on, so
		// a CRM-less deployment carries no crm leg anywhere (zero
		// config = byte-identical probes). The authenticated
		// HealthCheck is NON-critical by the seam's own doctrine (a
		// dead CRM costs toasts and raw names, never call flow).
		//samber-linter:allow hw-4 eagerly invoked below for probe discovery
		do.ProvideNamed(injector, "crm", func(i do.Injector) (*crm.Resolver, error) {
			return do.MustInvoke[*crm.Resolver](i), nil
		})
	}

	// Optional push seam (ring-while-closed): each transport is an
	// independent both-or-neither config; the fanout drops dead senders
	// at construction, so zero configured transports = a disabled fanout
	// (device API 404s, ring hook acknowledges inertly) and the product
	// stays byte-identical. A PRESENT but unusable configuration (bad
	// FCM credentials file, unparseable ntfy URL) fails the boot here —
	// the same doctrine as the passkey/Paperless seams.
	pushFanout, err := func() (*push.Fanout, error) {
		senders := map[push.Transport]push.Sender{}
		if cfg.Push.Ntfy.Enabled() {
			sender, err := push.NewNtfySender(cfg.Push.Ntfy.URL, cfg.Push.Ntfy.Token)
			if err != nil {
				return nil, err
			}
			senders[push.TransportNtfy] = sender
		}
		if cfg.Push.Fcm.Enabled() {
			sender, err := push.NewFcmSender(cfg.Push.Fcm.CredentialsFile)
			if err != nil {
				return nil, err
			}
			senders[push.TransportFcm] = sender
		}
		return push.NewFanout(senders, nil), nil
	}()
	if err != nil {
		return nil, wrapf("open push senders: %w", err)
	}
	do.Provide(injector, func(i do.Injector) (*push.Fanout, error) { return pushFanout, nil })

	// Optional passkey identity layer (embedded usermgmt): config-absent
	// → the service is never registered and Deps.UserAuth stays nil
	// (every passkey surface 404s). A present but UNUSABLE configuration
	// fails the boot at the eager invoke below — a login mode must never
	// half-exist (same doctrine as the Paperless archiver).
	if cfg.Auth.Passkey.Enabled() {
		//samber-linter:allow hw-4 eagerly resolved via InvokeNamed in New
		do.ProvideNamed(injector, "userauth", func(i do.Injector) (*userauth.Service, error) {
			bootCtx, cancel := context.WithTimeout(context.Background(), userauthBootTimeout)
			defer cancel()
			return userauth.New(bootCtx, PasskeyRuntime(cfg.Auth.Passkey), cfg.DataDir,
				do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).EnrollSvc(), log)
		})
	}

	// Optional OIDC/SSO login mode (Pocket ID): config-absent → the
	// provider is never registered and Deps.SSO stays nil (the login card
	// renders no SSO entry and the begin/callback routes 404). A present
	// but INVALID shape fails the boot at the eager invoke below; an
	// unreachable IdP does NOT (discovery is lazy — the phone must keep
	// working while the IdP is down).
	if cfg.Auth.OAuth2.Enabled() {
		//samber-linter:allow hw-4 eagerly resolved via InvokeNamed in New
		do.ProvideNamed(injector, "sso", func(i do.Injector) (*sso.Client, error) {
			return sso.New(SSOConfig(cfg.Auth))
		})
	}

	// Optional Paperless-ngx archive for inbound faxes: config-absent →
	// a nil fax.Archiver (fax.Service treats nil as off). A present but
	// UNUSABLE configuration fails the boot here — archiving must never
	// half-exist.
	archiver, err := paperless.NewArchiver(cfg.Paperless.URL, cfg.Paperless.Token, log)
	if err != nil {
		return nil, wrapf("paperless: %w", err)
	}

	// Optional supervised text-agent review seam: config-absent → a nil
	// client (no surface renders). A present but UNUSABLE configuration
	// (missing/empty secret file) fails the boot here — the seam must
	// never half-exist (same doctrine as the archiver above).
	var agent *textagent.Client
	if cfg.TextAgent.Enabled() {
		agent, err = textagent.New(cfg.TextAgent.URL, cfg.TextAgent.SecretFile)
		if err != nil {
			return nil, wrapf("textagent: %w", err)
		}
	}

	// --- services --------------------------------------------------------
	do.Provide(injector, func(i do.Injector) (*session.SQLiteStore, error) {
		return session.NewSQLiteStore(do.MustInvokeNamed[*store.Database](i, "sqlite").SQL(), cfg.SessionTTL)
	})
	//samber-linter:allow hw-1 health is the hand-assembled go-health probe (per-surface grading); the hub's liveness is the /events surface, never a do-healthcheck row
	do.Provide(injector, func(i do.Injector) (*server.ExtensionHubs, error) {
		return server.NewHubs(), nil
	})
	do.Provide(injector, func(i do.Injector) (*server.Notifier, error) {
		return server.NewNotifier(
			do.MustInvoke[*server.ExtensionHubs](i),
			do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).ThreadsSvc(),
			do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).FaxesSvc(), //nolint:branching-flow // composition-root resolution (AGENTS.md): the named cqrs service's fax backend (SUPERB Phase 6)
			do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).ContactsSvc(),
			do.MustInvoke[*crm.Resolver](i),
			cfg.ASR.Enabled(),
			views.FaxCapability{Send: cfg.Fax.SendEnabled(), Receive: cfg.Fax.ReceiveEnabled()},
		), nil
	})
	do.Provide(injector, func(i do.Injector) (gateway.MessageGateway, error) {
		return gateway.NewMessageGateway(cfg.Gateway, gateway.DefaultClient()), nil
	})
	do.Provide(injector, func(i do.Injector) (gateway.FaxGateway, error) {
		return gateway.NewFaxGateway(cfg.Gateway, gateway.DefaultClient()), nil
	})
	do.Provide(injector, func(i do.Injector) (*messaging.Service, error) {
		return messaging.New(
			do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).ThreadsSvc(),
			do.MustInvokeNamed[*blob.Store](i, "blob-dir"),
			do.MustInvoke[gateway.MessageGateway](i),
			do.MustInvoke[*server.Notifier](i).Project,
			cfg.Identities,
		), nil
	})
	do.Provide(injector, func(i do.Injector) (*fax.Service, error) {
		return fax.New(
			do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).FaxesSvc(), //nolint:branching-flow // composition-root resolution (AGENTS.md): the named cqrs service's fax backend (SUPERB Phase 6)
			do.MustInvokeNamed[*blob.Store](i, "blob-dir"),
			do.MustInvoke[gateway.FaxGateway](i),
			do.MustInvoke[*server.Notifier](i).Project,
			cfg.Identities,
			archiver,
		), nil
	})

	// --- probe + HTTP layer ---------------------------------------------
	// One probe instance, built once and threaded to both consumers
	// (server.Deps for /livez + /startupz, the dashboard for its cards)
	// — rebuilding it per consumer would fork the readiness truth.
	probe := health.New(injector,
		health.WithCriticalServices("sqlite", "blob-dir"),
		health.WithRefreshInterval(probeRefreshIf(cfg.Dashboard.Enable)),
	)

	// Flight recorder: nil unless the operator opted in. A boot failure
	// here is fatal on purpose — a tracing request that silently did not
	// happen would waste exactly the debugging session it exists for.
	trace, err := traceRecorderFor(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	if trace != nil {
		// The opt-in's observable state: an operator looking for the
		// seam in the journal finds WHERE snapshots land and WHAT the
		// capture policy is, without reading source.
		log.Info("flight recorder enabled (WEBPHONE_DEBUG_FLIGHTRECORDER=1)",
			"dir", traceDir(cfg.DataDir),
			"slow_request", traceSlowRequest.String(),
			"retention", traceRetention,
		)
	}

	var dashboardHandler http.Handler
	if cfg.Dashboard.Enable {
		dashboardHandler = newDashboard(injector, probe, cfg)
		// The dashboard ships no app-side authentication BY DECISION
		// (docs/planning/2026-11-04_23-45_dashboard-fence-decision.md):
		// behind any proxy the app cannot see the client address, and a
		// header gate would duplicate the module's fence mode as a
		// second config home. The fence is the reverse proxy's; say so
		// where the operator is looking.
		log.Warn("health dashboard mounted WITHOUT app-side authentication; the fence is your reverse proxy (module default: SSO gate in forwardAuth mode, loopback-only otherwise)")
	}

	do.Provide(injector, func(i do.Injector) (http.Handler, error) {
		db := do.MustInvokeNamed[*store.Database](i, "sqlite")
		blobs := do.MustInvokeNamed[*blob.Store](i, "blob-dir")
		var auth *userauth.Service
		if cfg.Auth.Passkey.Enabled() {
			auth = do.MustInvokeNamed[*userauth.Service](i, "userauth")
		}
		// SSO keeps a TRUE nil interface when the mode is off: a typed
		// nil *sso.Client assigned into the SSOClient interface reads
		// as non-nil to `deps.SSO != nil` — the routes would register,
		// the login card would render the SSO block (hiding the
		// extension form behind a closed disclosure), and the first
		// begin would panic on the nil receiver (the 2026-10-23
		// browser-suite M11 break, proven against the config-less
		// binary).
		var ssoClient serverauth.SSOClient
		if cfg.Auth.OAuth2.Enabled() {
			ssoClient = do.MustInvokeNamed[*sso.Client](i, "sso")
		}
		// The admin surface's identity inventory: config-derived, built
		// from the SAME shared mapping the login modes read (see
		// AdminService). A nil service (no admin configured on a mapping
		// that cannot mint logins) keeps every admin route on the styled
		// 404 — byte-identical to a deployment without the feature.
		adminService, err := AdminService(cfg)
		if err != nil {
			return nil, err
		}
		return server.New(server.Deps{
			Config:   cfg,
			Sessions: do.MustInvoke[*session.SQLiteStore](i),
			Messages: do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).ThreadsSvc(),
			Faxes:    do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).FaxesSvc(), //nolint:branching-flow // composition-root resolution (AGENTS.md): the named cqrs service's fax backend (SUPERB Phase 6)
			// The Phase-4 seams answer from the journal's decisions
			// (SUPERB): the journal is the one implementation behind their
			// readmodels gate interfaces (the engine stores retired with
			// the 2026-11-11 storage-ladder closeout).
			Contacts:    do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).ContactsSvc(),
			Snippets:    do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).SnippetsSvc(), //nolint:branching-flow // composition-root resolution (AGENTS.md): the named cqrs service's snippet backend (SUPERB Phase 3)
			Transcripts: do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).TranscriptsSvc(),
			Notes:       do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).NotesSvc(),
			Reviews:     do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).ReviewsSvc(),
			// The dial gate answers from the journal's decisions since the
			// SUPERB Phase 2 swap (the DialGate seam's one home is the
			// journal backend; the engine store retired).
			Dials: do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).Dials(),
			// The agent-audit store answers from the journal since the
			// SUPERB Phase 3 swap (the AgentAuditGate seam's one home is
			// the journal backend; the engine store retired).
			AgentAudit:  do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).AgentAuditSvc(),
			Messaging:   do.MustInvoke[*messaging.Service](i),
			Fax:         do.MustInvoke[*fax.Service](i),
			PhoneAPI:    do.MustInvoke[*pbx.Client](i),
			Hubs:        do.MustInvoke[*server.ExtensionHubs](i),
			Notifier:    do.MustInvoke[*server.Notifier](i),
			Shared:      cfg.Contacts,
			CRM:         do.MustInvoke[*crm.Resolver](i),
			ASR:         do.MustInvoke[asr.Provider](i),
			ASRLocality: asrLocality(cfg),
			TextAgent:   agent,
			DB:          db.SQL(),
			BlobRoot:    blobs.Root(),
			Probe:       probe,
			Dashboard:   dashboardHandler,
			UserAuth:    auth,
			SSO:         ssoClient,
			Trace:       traceSeam(trace),
			// The snapshot counter family rides the adapter: non-nil
			// exactly when the recorder seam is on.
			TraceMetrics: traceMetricsSeam(trace),
			// The dedup window lives in system.db since A43 (sys.Dedup, the
			// same CAS capability the fax replay guard rides): readmodels.db's
			// window retired with the engine.
			Dedup: engineDedupWindow{store: systemDedup(i)},
			Push:  pushFanout,
			// The device registry answers from the journal since the SUPERB
			// Phase 3 swap (the PushDeviceGate seam's one home is the
			// journal backend; the engine store retired).
			PushDevices: do.MustInvokeNamed[*CqrsService](i, cqrsProviderName).PushDevicesSvc(),
			Admin:       adminService,
		}), nil
	})

	// Eager instantiation, fail fast with named errors. The critical
	// pair must exist before any probe batch runs (a never-invoked lazy
	// service health-checks as silently passing); the handler pulls the
	// whole service graph, so its invocation surfaces every remaining
	// constructor failure here instead of at first request.
	if _, err := do.InvokeNamed[*store.Database](injector, "sqlite"); err != nil {
		return nil, wrapf("open store: %w", err)
	}
	if _, err := do.InvokeNamed[*blob.Store](injector, "blob-dir"); err != nil {
		return nil, wrapf("open blob store: %w", err)
	}
	if _, err := do.InvokeNamed[*CqrsService](injector, cqrsProviderName); err != nil {
		return nil, wrapf("open cqrs system: %w", err)
	}
	// The read-model engine and its one-time importers are GONE (A43):
	// every seam answers from system.db, and readmodels.db stopped opening
	// at boot. Upgrades ride the v2.10.0/v2.11.0 import trains first (the
	// retirement plan doc names the boundary).

	if cfg.Auth.Passkey.Enabled() {
		if _, err := do.InvokeNamed[*userauth.Service](injector, "userauth"); err != nil {
			return nil, wrapf("open passkey auth: %w", err)
		}
	}
	if cfg.Auth.OAuth2.Enabled() {
		if _, err := do.InvokeNamed[*sso.Client](injector, "sso"); err != nil {
			return nil, wrapf("wire oidc login: %w", err)
		}
	}
	if cfg.ASR.Enabled() {
		if _, err := do.InvokeNamed[asr.Provider](injector, "asr"); err != nil {
			return nil, wrapf("open asr provider: %w", err)
		}
	}
	handler, err := do.Invoke[http.Handler](injector)
	if err != nil {
		return nil, wrapf("wire server: %w", err)
	}

	//nolint:branching-flow // MustInvoke in the composition root is doctrine (DO-1 fail-fast miswires); arch-test pins it to internal/app
	return &App{injector: injector, cfg: cfg, log: log, Probe: probe, Trace: trace, Handler: handler, Hubs: do.MustInvoke[*server.ExtensionHubs](injector), Push: pushFanout}, nil
}

// probeRefreshIf keeps the probe in live mode (interval 0, the exact
// pre-container behavior) unless the dashboard is mounted — the pusher
// needs the background cache loop to observe changes.
func probeRefreshIf(dashboardEnabled bool) time.Duration {
	if dashboardEnabled {
		return probeRefresh
	}
	return 0
}

// asrLocality derives the seam's config-honest trust label ("local" |
// "remote", asr.LocalityOf); empty when the seam is off (no label may
// imply processing that does not happen).
func asrLocality(cfg config.Config) string {
	if !cfg.ASR.Enabled() {
		return ""
	}
	return asr.LocalityOf(cfg.ASR.URL)
}

// passkeyRuntime adapts the validated config into the userauth runtime
// shape (extensions parsed once, at boot — config validation has already
// rejected unparseable ones, so the parse failures here are impossible
// by construction and panic honestly on a broken invariant).
// PasskeyRuntime is the config→runtime adapter for the embedded
// identity layer (the arch invariant keeps userauth config-free, so
// THIS package owns the translation). Exported for the CLI enrollment
// path, which builds the same runtime the server boots with.
func PasskeyRuntime(p config.Passkey) userauth.PasskeyRuntime {
	users := make(map[string]userauth.MappedUser, len(p.Users))
	for email, user := range p.Users {
		extensions := make([]domain.Extension, 0, len(user.Extensions))
		for _, raw := range user.Extensions {
			ext, err := domain.ParseExtension(raw)
			if err != nil {
				panic(fmt.Sprintf("passkey mapping for %q holds unparseable extension %q after config validation: %v", email, raw, err)) //cqrs-lint:ignore(C009) post-validation impossible state: fail loudly at boot (DO-1)
			}
			extensions = append(extensions, ext)
		}
		users[email] = userauth.MappedUser{Email: email, DisplayName: user.DisplayName, Extensions: extensions}
	}
	return userauth.PasskeyRuntime{
		RPID:                   p.RPID,
		RPDisplayName:          p.RPDisplayName,
		RPOrigins:              p.RPOrigins,
		Users:                  users,
		ExtensionPasswordFiles: p.ExtensionPasswordFiles,
	}
}

// SSOConfig is the config→runtime adapter for the OIDC login mode. The
// mapping is the SHARED one (auth.users, folded into the passkey shape by
// config.Load) so passkey and SSO can never disagree about who may sign
// in or where an extension's SIP password lives.
func SSOConfig(a config.Auth) sso.Config {
	users := a.Users
	if len(users) == 0 {
		users = a.Passkey.Users
	}
	passwordFiles := a.ExtensionPasswordFiles
	if len(passwordFiles) == 0 {
		passwordFiles = a.Passkey.ExtensionPasswordFiles
	}
	mapped := make(map[string]authmap.User, len(users))
	for email, user := range users {
		extensions := make([]domain.Extension, 0, len(user.Extensions))
		for _, raw := range user.Extensions {
			ext, err := domain.ParseExtension(raw)
			if err != nil {
				//nolint:branching-flow // composition-root fail-fast (DO-1): config validation already rejects unparseable extensions, so reaching here is an unrepresentable miswire
				panic(fmt.Sprintf("auth mapping for %q holds unparseable extension %q after config validation: %v", email, raw, err)) //cqrs-lint:ignore(C009) post-validation impossible state: fail loudly at boot (DO-1)
			}
			extensions = append(extensions, ext)
		}
		mapped[email] = authmap.User{Email: email, DisplayName: user.DisplayName, Extensions: extensions}
	}
	return sso.Config{
		Issuer:                 a.OAuth2.Issuer,
		ClientID:               a.OAuth2.ClientID,
		ClientSecretFile:       a.OAuth2.ClientSecretFile,
		RedirectURL:            a.OAuth2.RedirectURL,
		Scopes:                 a.OAuth2.Scopes,
		AllowedGroups:          a.OAuth2.AllowedGroups,
		Users:                  mapped,
		ExtensionPasswordFiles: passwordFiles,
	}
}

// AdminService is the config→runtime adapter for the admin surface: it
// builds the SAME shared mapping the login modes read (auth.users, with
// the historical passkey.users fallback) into the admin service's own
// authmap, plus the identities feed. Boot-compatibility doctrine: a
// mapping that cannot mint logins (a mapped extension without a
// password file) was already invalid for every login MODE — but with
// both modes off such a config historically booted, so it still does:
// the service degrades to nil unless an ADMIN is configured on the
// broken mapping (then the boot fails loudly; an admin that cannot
// exist is worse than a failed start).
func AdminService(cfg config.Config) (*admin.Service, error) {
	users := cfg.Auth.Users
	if len(users) == 0 {
		users = cfg.Auth.Passkey.Users
	}
	passwordFiles := cfg.Auth.ExtensionPasswordFiles
	if len(passwordFiles) == 0 {
		passwordFiles = cfg.Auth.Passkey.ExtensionPasswordFiles
	}
	anyAdmin := false
	mapped := make(map[string]authmap.User, len(users))
	for email, user := range users {
		extensions := make([]domain.Extension, 0, len(user.Extensions))
		for _, raw := range user.Extensions {
			ext, err := domain.ParseExtension(raw)
			if err != nil {
				//nolint:branching-flow // composition-root fail-fast (DO-1): config validation already rejects unparseable extensions, so reaching here is an unrepresentable miswire
				panic(fmt.Sprintf("auth mapping for %q holds unparseable extension %q after config validation: %v", email, raw, err)) //cqrs-lint:ignore(C009) post-validation impossible state: fail loudly at boot (DO-1)
			}
			extensions = append(extensions, ext)
		}
		anyAdmin = anyAdmin || user.Admin
		mapped[email] = authmap.User{Email: email, DisplayName: user.DisplayName, Extensions: extensions, Admin: user.Admin}
	}
	authMapped, err := authmap.New(mapped, passwordFiles)
	if err != nil {
		if anyAdmin {
			return nil, wrapf("wire admin surface: %w", err)
		}
		return nil, nil
	}
	return admin.New(authMapped, cfg.Identities)
}

// Start launches the background loops: the dashboard pusher (when
// mounted), the probe's refresh loop (when cached mode), and the
// bounded retention sweep. ctx bounds them all — cancel it and they end.
func (a *App) Start(ctx context.Context) error {
	if a.cfg.Dashboard.Enable {
		dash, err := do.InvokeNamed[*dashboard.Dashboard](a.injector, "dashboard") //nolint:branching-flow // composition-root resolution (AGENTS.md): Start fail-fast resolves the named dashboard; no business logic touches the container
		if err != nil {
			return wrapf("resolve dashboard: %w", err)
		}
		if err := dash.Start(ctx); err != nil {
			return wrapf("start dashboard pusher: %w", err)
		}
	}
	if err := a.Probe.Start(ctx); err != nil {
		return wrapf("start health probe: %w", err)
	}
	db, err := do.InvokeNamed[*store.Database](a.injector, "sqlite") //nolint:branching-flow // composition-root resolution (AGENTS.md): the named critical pair's sqlite leg
	if err != nil {
		return wrapf("resolve store: %w", err)
	}
	blobs, err := do.InvokeNamed[*blob.Store](a.injector, "blob-dir") //nolint:branching-flow // composition-root resolution (AGENTS.md): the named critical pair's blob-dir leg
	if err != nil {
		return wrapf("resolve blob store: %w", err)
	}
	// The system's projection workers and timers start BEFORE the HTTP
	// listener accepts traffic (contract §3), so the initial journal
	// drain never races a request that expects a caught-up view.
	cqrs, err := do.InvokeNamed[*CqrsService](a.injector, cqrsProviderName) //nolint:branching-flow // composition-root resolution (AGENTS.md): the named cqrs service
	if err != nil {
		return wrapf("resolve cqrs system: %w", err)
	}
	if err := cqrs.Start(ctx); err != nil {
		return wrapf("start cqrs system: %w", err)
	}
	// Arm the fanout's prune registry now that the push-devices store
	// resolves (it rides the cqrs system): start order guarantees every
	// later Fan sees an armed registry, and the adapter keeps the fanout
	// dependency-light (domain types only).
	if a.Push != nil {
		a.Push.WithRegistry(pushRegistryAdapter{devices: cqrs.PushDevicesSvc()})
	}
	retention.Start(ctx, db.SQL(), blobs, retention.JournalSweepers{
		Threads: cqrs.ThreadsSvc(), Faxes: cqrs.FaxesSvc(), Transcripts: cqrs.TranscriptsSvc(),
	}, time.Duration(a.cfg.RetentionDays)*24*time.Hour)
	startDedupSweeper(ctx, a.log, engineDedupWindow{store: systemDedup(a.injector)})
	// The chat-network lanes wire here (config-gated; absent config wires
	// nothing) and run under the same ctx bound: a lane exiting while the
	// app lives is a loud Error, not silent intake death.
	loops, err := wireChatNetworks(a.cfg, a.injector, db.SQL(), a.log)
	if err != nil {
		return wrapf("wire chat networks: %w", err)
	}
	for _, loop := range loops {
		go func(run func(context.Context) error) {
			if err := run(ctx); err != nil && ctx.Err() == nil {
				a.log.Error("chat receiver loop exited", "error", err)
			}
		}(loop)
	}
	return nil
}

// Shutdown drains the application: the probe marks itself shutting
// down first (readiness flips while the HTTP drain may still be
// serving), then the container cascade closes every Shutdowner —
// the dashboard's pusher and the SQLite handle among them. Idempotent
// per service; a second call is inert.
func (a *App) Shutdown() error {
	a.Probe.Shutdown()
	report := a.injector.Shutdown()
	// The flight recorder leaves the container untouched (see App.Trace),
	// so its drain is explicit — last, after every handler that could
	// still report a request has stopped.
	if a.Trace != nil {
		if err := a.Trace.Close(); err != nil {
			a.log.Warn("flight recorder close failed", "error", err)
		}
	}
	if report == nil || report.Succeed {
		return nil
	}
	errs := make([]error, 0, len(report.Errors))
	for _, err := range report.Errors {
		errs = append(errs, err)
	}
	return &ShutdownError{errs: errs}
}

// ShutdownError aggregates the container's per-service shutdown
// failures. The aggregate stays family-neutral — every wrapped cause
// carries its own code and family — and it traverses like errors.Join's
// product (Unwrap() []error, one line per cause), so sentinel matching
// and family classification see every half.
type ShutdownError struct {
	errs []error
}

func (e *ShutdownError) Error() string {
	switch len(e.errs) {
	case 0:
		return ""
	case 1:
		return e.errs[0].Error()
	}
	var b strings.Builder
	for i, err := range e.errs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (e *ShutdownError) Unwrap() []error { return e.errs }
