// Command webphone serves the unified-communications web app: the SIP call
// island, SMS/MMS threads, fax, voicemail, history, contacts and settings —
// one binary, one page, one login against the PBX directory.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larsartmann/httputil"

	"github.com/larsartmann/webphone/internal/app"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/userauth"
)

// main reports and exits: run()'s designed errors render the five-part
// operator contract and exit 1 (bootreport.go); a panic escaping run()
// renders the block plus the trace and exits 2. Taxonomy + copy:
// docs/error-contract.md, "Boot surface".
func main() {
	if err := run(); err != nil {
		reportBootFailure(err)
	}
}

// propagatef is the ONE home for startup wiring's family-neutral
// wraps (family-adoption train, 2026-09-30): each run() step wraps an
// inner error whose family varies by cause (config mistakes are
// Rejections, storage failures Infrastructure), so a fixed-family
// wrap here would clobber the constructor's classification. The inner
// error owns the family; this only adds context.
func propagatef(format string, args ...any) error {
	return fmt.Errorf(format, args...) //nolint:erraudit // family-neutral propagation: the inner error owns the family
}

// enrollPasskey serves the `-enroll-passkey <email>` one-shot: it must
// run under the SAME config and data dir as the server (the sqlite
// identity store lives there — as the webphone service user on a
// deployment), registers the account idempotently and mints the
// one-time token. The printed URL names the first configured origin;
// the token itself is single-use and expires after
// userauth.EnrollTokenTTL.
func enrollPasskey(email string) error {
	cfg, err := config.Load()
	if err != nil {
		return bootErr(bootConfigInvalid, propagatef("load config: %w", err))
	}
	if !cfg.Auth.Passkey.Enabled() {
		return bootErr(bootConfigInvalid, errors.New("passkey mode is not configured (auth.passkey.*) — enable and map the email before enrolling")) //nolint:erraudit // boot root cause: bootErr owns the 5-part operator surface; no inner family to propagate
	}
	if _, mapped := cfg.Auth.Passkey.Users[email]; !mapped {
		return bootErr(bootConfigInvalid, fmt.Errorf("email %q is not mapped in auth.passkey.users — add it to the config first", email)) //nolint:erraudit // boot root cause: bootErr owns the 5-part operator surface; no inner family to propagate
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	svc, err := userauth.New(ctx, app.PasskeyRuntime(cfg.Auth.Passkey), cfg.DataDir, slog.Default())
	if err != nil {
		return propagatef("open identity store: %w", err)
	}
	defer func() {
		if err := svc.Shutdown(); err != nil {
			slog.Warn("identity store shutdown failed after enrollment", "error", err)
		}
	}()
	userID, err := svc.Register(ctx, email)
	if err != nil {
		return propagatef("register %s: %w", email, err)
	}
	token, err := svc.MintEnrollToken(ctx, email, userID)
	if err != nil {
		return propagatef("mint enrollment token: %w", err)
	}
	fmt.Printf("Passkey enrollment link for %s (single use, expires after %s):\n\n  %s/enroll?token=%s\n\nOpen it in a browser on a device you carry, then follow the passkey prompt.\n", email, userauth.EnrollTokenTTL, cfg.Auth.Passkey.RPOrigins[0], token)
	return nil
}

func run() error {
	// First-registered defer runs last: a panic from ANY boot step (or
	// from the defers below) still lands in the operator report.
	defer recoverBootPanic()
	// The one-shot enrollment surface runs BEFORE any server wiring:
	// `webphone -enroll-passkey lars@example.com` opens the same
	// identity store the server uses, registers the account if needed
	// and prints a 15-minute one-time enrollment link (D5: the operator
	// is the identity authority; no self-serve email round-trip).
	if len(os.Args) >= 3 && os.Args[1] == "-enroll-passkey" {
		return enrollPasskey(os.Args[2])
	}
	cfg, err := config.Load()
	if err != nil {
		return bootErr(bootConfigInvalid, propagatef("load config: %w", err))
	}
	slog.Info("webphone starting",
		"addr", cfg.Addr, "dataDir", cfg.DataDir, "gateway", string(cfg.Gateway.Mode))
	// Operators must SEE the fronting shape at boot: an empty list pair is
	// the exact configuration that 403s every browser login behind a
	// TLS-terminating proxy (hosts only, never secret material).
	slog.Info("csrf fronting",
		"trustedProxies", len(cfg.CSRF.TrustedProxies), "trustedOrigins", cfg.CSRF.TrustedOrigins)
	// Login honesty: with a phone API configured, POST /api/session
	// verifies the submitted credentials against the PBX directory and
	// fails closed (401/502). Without one (loopback dev), logins are
	// unverified by construction — operators must see that state.
	if cfg.PhoneAPIURL == "" {
		slog.Warn("pbx credential verification disabled (no phone_api_url configured; logins are not verified)")
	} else {
		slog.Info("pbx credential verification", "mode", "enforced")
	}
	// CRM integration visibility: off is the default and fine; on means
	// caller names in tabs and post-call journal entries in the CRM.
	if cfg.CRM.URL != "" {
		slog.Info("crm integration", "mode", "enabled", "url", cfg.CRM.URL)
	}
	if cfg.Dashboard.Enable {
		slog.Info("health dashboard", "mode", "enabled", "path", "/health")
	}
	// Passkey mode visibility: on means the login card's front door is
	// the email + WebAuthn ceremony (extension login stays as
	// break-glass); off is the default and logs nothing — the login
	// card is byte-shaped like the pre-passkey one.
	if cfg.Auth.Passkey.Enabled() {
		slog.Info("passkey auth",
			"mode", "enabled", "rpid", cfg.Auth.Passkey.RPID, "users", len(cfg.Auth.Passkey.Users))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The container (internal/app) owns object lifetime: every service,
	// the go-health probe over the container's health-checkable
	// services, and the optional dashboard.
	application, err := app.New(cfg, slog.Default())
	if err != nil {
		return propagatef("build app: %w", err)
	}
	if err := application.Start(ctx); err != nil {
		return bootErr(bootGeneric, propagatef("start app: %w", err))
	}

	// httputil.Server owns the lifecycle — the same tested primitive the
	// cqrs-htmx setup bundle's RunHandler wraps (footprint train
	// 2026-09-30: the setup ADOPTION measured +10.4 MB / +68.2% binary
	// delta and failed the recorded go/no-go gate, so the bundle stays
	// rejected; the lifecycle value rides the wrapper's own dependency).
	// SSE-safe timeout set: ReadHeaderTimeout bounds slowloris, IdleTimeout
	// reaps dead keep-alives, NO Read/Write deadlines — SSE streams outlive
	// any fixed deadline. The 30s shutdown budget covers the SSE hub drain.
	httpServer, err := httputil.NewServer(httputil.ServerConfig{
		Addr:              cfg.Addr,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   30 * time.Second,
	}, application.Handler)
	if err != nil {
		return errors.Join(bootErr(bootListen, err), application.Shutdown()) //nolint:erraudit // aggregate of independently-stamped errors; errorfamily has no Join
	}

	slog.Info("listening", "addr", cfg.Addr)
	errCh := httpServer.Start()

	select {
	case err := <-errCh:
		return errors.Join(bootErr(bootListen, propagatef("serve: %w", err)), application.Shutdown()) //nolint:erraudit // aggregate of independently-stamped errors; errorfamily has no Join
	case <-ctx.Done():
		slog.Info("shutting down")
		// Order matters: the HTTP drain finishes in-flight responses
		// (SSE hubs flush), THEN the container closes the dashboard
		// pusher and the SQLite handle.
		if err := httpServer.Shutdown(context.WithoutCancel(ctx)); err != nil {
			return errors.Join(propagatef("http shutdown: %w", err), application.Shutdown()) //nolint:erraudit // aggregate of independently-stamped errors; errorfamily has no Join
		}
		return application.Shutdown()
	}
}
