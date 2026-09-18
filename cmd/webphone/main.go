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
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/fax"
	"github.com/larsartmann/webphone/internal/gateway"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/pbx"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	// --- services ----------------------------------------------------------
	messages := store.NewMessages(db)
	faxes := store.NewFaxes(db)
	contacts := store.NewContacts(db)
	sessions := session.NewStore(cfg.SessionTTL)

	hubs := server.NewHubs()
	notifier := server.NewNotifier(hubs, messages, faxes)

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
		Shared:    sharedContacts(cfg),
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

func sharedContacts(cfg config.Config) []domain.SharedContact {
	return cfg.Contacts
}
