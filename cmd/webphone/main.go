// Command webphone serves the unified-communications web app: the SIP call
// island, SMS/MMS threads, fax, voicemail, history, contacts and settings —
// one binary, one page, one login against the PBX directory.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

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

func main() {
	if err := run(); err != nil {
		slog.Error("webphone exited", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create the data dir up front: a missing parent directory would
	// otherwise surface as SQLite's cryptic "unable to open database
	// file (14)".
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// --- infrastructure --------------------------------------------------
	db, err := store.Open(filepath.Join(cfg.DataDir, "webphone.db"))
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer func() { _ = db.Close() }()

	blobs, err := blob.New(filepath.Join(cfg.DataDir, "files"))
	if err != nil {
		return fmt.Errorf("open blob store: %w", err)
	}

	phoneAPI, err := pbx.NewClient(cfg.PhoneAPIURL)
	if err != nil {
		return fmt.Errorf("phone api client: %w", err)
	}

	crmClient, err := crm.NewClient(cfg.CRM.URL, cfg.CRM.Token)
	if err != nil {
		return fmt.Errorf("crm client: %w", err)
	}
	crmResolver := crm.NewResolver(crmClient, slog.Default())

	// --- services ----------------------------------------------------------
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
	contacts := store.NewContacts(db)
	sessions, err := session.NewSQLiteStore(db, cfg.SessionTTL)
	if err != nil {
		return fmt.Errorf("open session store: %w", err)
	}

	// Bounded retention (plan T25): retention_days > 0 starts the daily
	// sweep; the default keeps everything forever and starts nothing.
	retention.Start(ctx, db, blobs, time.Duration(cfg.RetentionDays)*24*time.Hour)

	hubs := server.NewHubs()
	notifier := server.NewNotifier(hubs, messages, faxes, crmResolver)

	messageGateway := gateway.NewMessageGateway(cfg.Gateway, gateway.DefaultClient())
	faxGateway := gateway.NewFaxGateway(cfg.Gateway, gateway.DefaultClient())

	messagingService := messaging.New(messages, blobs, messageGateway, notifier.MessagesChanged)
	faxService := fax.New(faxes, blobs, faxGateway, notifier.FaxChanged)

	handler := server.New(server.Deps{
		Config:    cfg,
		Sessions:  sessions,
		Messages:  messages,
		Faxes:     faxes,
		Contacts:  contacts,
		Messaging: messagingService,
		Fax:       faxService,
		PhoneAPI:  phoneAPI,
		Hubs:      hubs,
		Shared:    cfg.Contacts,
		CRM:       crmResolver,
		DB:        db,
		BlobRoot:  blobs.Root(),
	})

	// --- serve ---------------------------------------------------------------
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		// no WriteTimeout: SSE streams outlive any fixed deadline
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}
