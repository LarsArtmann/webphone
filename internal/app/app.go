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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/larsartmann/go-health"
	dashboard "github.com/larsartmann/go-health-dashboard"
	"github.com/samber/do/v2"

	"github.com/larsartmann/webphone/internal/blob"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
	"github.com/larsartmann/webphone/internal/retention"
	"github.com/larsartmann/webphone/internal/server"
	"github.com/larsartmann/webphone/internal/session"
	"github.com/larsartmann/webphone/internal/store"
)

// Adapter-side lifecycle guards (the service packages stay free of
// framework imports; the container contract is asserted where the
// container lives). store.Database additionally carries the shutdown
// cascade for the SQLite handle.
var (
	_ do.HealthcheckerWithContext = (*store.Database)(nil)
	_ do.HealthcheckerWithContext = (*blob.Store)(nil)
	_ do.ShutdownerWithError      = (*store.Database)(nil)
	_ do.Shutdowner               = (*dashboard.Dashboard)(nil)
)

// probeRefresh runs the probe's background cache loop while the
// dashboard is enabled: its pusher reads CachedResponse on every tick,
// which in live mode (interval 0) would forever hold the boot snapshot.
// Two local checks per second are noise. Disabled dashboard → interval
// 0, exactly the pre-container live-mode behavior.
const probeRefresh = time.Second

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
	Probe   *health.Probe
	Handler http.Handler
}

// wrapf is this package's home for family-neutral startup wraps (same
// contract as cmd/webphone's propagatef): each step wraps an inner
// error whose family varies by cause, so a fixed-family wrap here would
// clobber the constructor's classification.
func wrapf(format string, args ...any) error {
	return fmt.Errorf(format, args...) //nolint:erraudit // family-neutral propagation: the inner error owns the family
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
	// (health.WithCriticalServices below) — the names are contract.
	do.ProvideNamed(injector, "sqlite", func(i do.Injector) (*store.Database, error) {
		return store.OpenDatabase(filepath.Join(cfg.DataDir, "webphone.db"))
	})
	do.ProvideNamed(injector, "blob-dir", func(i do.Injector) (*blob.Store, error) {
		return blob.New(filepath.Join(cfg.DataDir, "files"))
	})
	do.Provide(injector, func(i do.Injector) (*pbx.Client, error) {
		return pbx.NewClient(cfg.PhoneAPIURL)
	})
	do.Provide(injector, func(i do.Injector) (*crm.Client, error) {
		return crm.NewClient(cfg.CRM.URL, cfg.CRM.Token)
	})
	do.Provide(injector, func(i do.Injector) (*crm.Resolver, error) {
		return crm.NewResolver(do.MustInvoke[*crm.Client](i), log), nil
	})

	// --- services --------------------------------------------------------
	do.Provide(injector, func(i do.Injector) (*store.Messages, error) {
		return store.NewMessages(do.MustInvokeNamed[*store.Database](i, "sqlite").SQL()), nil
	})
	do.Provide(injector, func(i do.Injector) (*store.Faxes, error) {
		return store.NewFaxes(do.MustInvokeNamed[*store.Database](i, "sqlite").SQL()), nil
	})
	do.Provide(injector, func(i do.Injector) (*store.Contacts, error) {
		return store.NewContacts(do.MustInvokeNamed[*store.Database](i, "sqlite").SQL()), nil
	})
	do.Provide(injector, func(i do.Injector) (*session.SQLiteStore, error) {
		return session.NewSQLiteStore(do.MustInvokeNamed[*store.Database](i, "sqlite").SQL(), cfg.SessionTTL)
	})
	do.Provide(injector, func(i do.Injector) (*server.ExtensionHubs, error) {
		return server.NewHubs(), nil
	})
	do.Provide(injector, func(i do.Injector) (*server.Notifier, error) {
		return server.NewNotifier(
			do.MustInvoke[*server.ExtensionHubs](i),
			do.MustInvoke[*store.Messages](i),
			do.MustInvoke[*store.Faxes](i),
			do.MustInvoke[*crm.Resolver](i),
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
			do.MustInvoke[*store.Messages](i),
			do.MustInvokeNamed[*blob.Store](i, "blob-dir"),
			do.MustInvoke[gateway.MessageGateway](i),
			do.MustInvoke[*server.Notifier](i).MessagesChanged,
			cfg.Identities,
		), nil
	})
	do.Provide(injector, func(i do.Injector) (*fax.Service, error) {
		return fax.New(
			do.MustInvoke[*store.Faxes](i),
			do.MustInvokeNamed[*blob.Store](i, "blob-dir"),
			do.MustInvoke[gateway.FaxGateway](i),
			do.MustInvoke[*server.Notifier](i).FaxChanged,
			cfg.Identities,
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

	var dashboardHandler http.Handler
	if cfg.Dashboard.Enable {
		dashboardHandler = newDashboard(injector, probe, cfg)
	}

	do.Provide(injector, func(i do.Injector) (http.Handler, error) {
		db := do.MustInvokeNamed[*store.Database](i, "sqlite")
		blobs := do.MustInvokeNamed[*blob.Store](i, "blob-dir")

		return server.New(server.Deps{
			Config:    cfg,
			Sessions:  do.MustInvoke[*session.SQLiteStore](i),
			Messages:  do.MustInvoke[*store.Messages](i),
			Faxes:     do.MustInvoke[*store.Faxes](i),
			Contacts:  do.MustInvoke[*store.Contacts](i),
			Messaging: do.MustInvoke[*messaging.Service](i),
			Fax:       do.MustInvoke[*fax.Service](i),
			PhoneAPI:  do.MustInvoke[*pbx.Client](i),
			Hubs:      do.MustInvoke[*server.ExtensionHubs](i),
			Shared:    cfg.Contacts,
			CRM:       do.MustInvoke[*crm.Resolver](i),
			DB:        db.SQL(),
			BlobRoot:  blobs.Root(),
			Probe:     probe,
			Dashboard: dashboardHandler,
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
	handler, err := do.Invoke[http.Handler](injector)
	if err != nil {
		return nil, wrapf("wire server: %w", err)
	}

	return &App{injector: injector, cfg: cfg, log: log, Probe: probe, Handler: handler}, nil
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

// Start launches the background loops: the dashboard pusher (when
// mounted), the probe's refresh loop (when cached mode), and the
// bounded retention sweep. ctx bounds them all — cancel it and they end.
func (a *App) Start(ctx context.Context) error {
	if a.cfg.Dashboard.Enable {
		dash := do.MustInvokeNamed[*dashboard.Dashboard](a.injector, "dashboard")
		dash.Start(ctx)
	}
	if err := a.Probe.Start(ctx); err != nil {
		return wrapf("start health probe: %w", err)
	}
	db := do.MustInvokeNamed[*store.Database](a.injector, "sqlite")
	blobs := do.MustInvokeNamed[*blob.Store](a.injector, "blob-dir")
	retention.Start(ctx, db.SQL(), blobs, time.Duration(a.cfg.RetentionDays)*24*time.Hour)
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
	if report == nil || report.Succeed {
		return nil
	}
	errs := make([]error, 0, len(report.Errors))
	for _, err := range report.Errors {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
