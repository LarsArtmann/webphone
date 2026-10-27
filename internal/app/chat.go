package app

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/larsartmann/go-error-family"
	"github.com/larsartmann/webphone/internal/config"
	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/matrix"
	"github.com/larsartmann/webphone/internal/messaging"
	"github.com/larsartmann/webphone/internal/signal"
	"github.com/larsartmann/webphone/internal/store"
	"github.com/samber/do/v2"
)

// chatBootTimeout bounds the boot-time reachability checks below; it
// is deliberately short — a slow homeserver must not stall boot, the
// receiver retries in the background anyway.
const chatBootTimeout = 10 * time.Second

// wireChatNetworks builds the optional chat-network lanes: senders
// register behind their config gate, receivers come back as loops the
// caller starts with the app context. Config-absent → nothing
// registers, nothing runs: the product stays byte-identical.
//
// Boot doctrine (mirrors SSO's lazy discovery): a WRONG configuration
// (bad token, refused identity) fails the boot — the lane must never
// half-exist; an UNREACHABLE peer (homeserver down, daemon not yet
// started) only WARNs — the phone keeps working and the receiver
// retries with backoff.
func wireChatNetworks(
	cfg config.Config, injector *do.RootScope, db *sql.DB, log *slog.Logger,
) ([]func(context.Context) error, error) {
	var loops []func(context.Context) error
	messagingSvc := do.MustInvoke[*messaging.Service](injector)

	if cfg.Matrix.Enabled() {
		client := matrix.New(cfg.Matrix.Homeserver, cfg.Matrix.UserID, cfg.Matrix.AccessToken)
		bootCtx, cancel := context.WithTimeout(context.Background(), chatBootTimeout)
		if err := client.Whoami(bootCtx); err != nil {
			cancel()
			if errorfamily.Classify(err) == errorfamily.Rejection {
				return nil, wrapf("matrix: %w", err)
			}
			log.Warn("matrix: homeserver unreachable at boot; the receiver retries in the background", "error", err)
		} else {
			cancel()
		}
		owner := domain.MustParseExtension(cfg.Matrix.Owner)
		if err := messagingSvc.RegisterSender(domain.NetworkMatrix, client); err != nil {
			return nil, wrapf("matrix: %w", err)
		}
		receiver := matrix.NewReceiver(client, owner, messagingSvc, appFlagCursor{db: db, key: "matrix.sync_cursor"}, log)
		loops = append(loops, receiver.Run)
		log.Info("matrix lane wired", "homeserver", cfg.Matrix.Homeserver, "user", cfg.Matrix.UserID, "owner", cfg.Matrix.Owner)
	}

	if cfg.Signal.Enabled() {
		client := signal.New(cfg.Signal.SocketPath, cfg.Signal.Account, log)
		owner := domain.MustParseExtension(cfg.Signal.Owner)
		if err := messagingSvc.RegisterSender(domain.NetworkSignal, client); err != nil {
			return nil, wrapf("signal: %w", err)
		}
		// The socket may legitimately not exist yet at boot (service
		// ordering); the receiver's ping loop reconnects with backoff,
		// so the boot check would only duplicate it. Wiring is enough.
		receiver := signal.NewReceiver(client, owner, messagingSvc, messagingSvc, log)
		loops = append(loops, receiver.Run)
		log.Info("signal lane wired", "socket", cfg.Signal.SocketPath, "account", cfg.Signal.Account, "owner", cfg.Signal.Owner)
	}

	return loops, nil
}

// appFlagCursor adapts the app_flags table onto the matrix sync
// cursor: one home (store.FlagValue/SetFlag), one key per lane.
type appFlagCursor struct {
	db  *sql.DB
	key string
}

func (c appFlagCursor) Load(ctx context.Context) (string, error) {
	value, _, err := store.FlagValue(ctx, c.db, c.key)
	return value, err
}

func (c appFlagCursor) Save(ctx context.Context, token string) error {
	return store.SetFlag(ctx, c.db, c.key, token)
}
