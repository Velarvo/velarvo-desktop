package app

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Velarvo/velarvo-desktop/internal/client"
	"github.com/Velarvo/velarvo-desktop/internal/keychain"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/projects"
	"github.com/Velarvo/velarvo-desktop/internal/session"
	"github.com/Velarvo/velarvo-desktop/internal/settings"
	"github.com/Velarvo/velarvo-desktop/internal/sshconn"
	"github.com/Velarvo/velarvo-desktop/internal/storage"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
	"github.com/Velarvo/velarvo-desktop/internal/workspaces"
	"github.com/awnumar/memguard"
)

type App struct {
	ctx        context.Context
	client     *client.Client
	keychain   *keychain.Service
	session    *session.Manager
	db         *storage.DB
	vault      *vault.Service
	projects   *projects.Service
	workspaces *workspaces.Service
	settings   *settings.Service
	ssh        *sshconn.Service
}

func New(c *client.Client, kc *keychain.Service) *App {
	return &App{
		client:   c,
		keychain: kc,
		session:  session.Default,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	memguard.CatchInterrupt()

	dataDir, err := userDataDir()
	if err != nil {
		logger.Named("app").Fatalw("failed to resolve user data directory", "error", err)
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		logger.Named("app").Fatalw("failed to create user data directory", "path", dataDir, "error", err)
	}

	db, err := storage.Open(dataDir)
	if err != nil {
		logger.Named("app").Fatalw("failed to initialize local storage", "path", dataDir, "error", err)
	}
	a.db = db

	a.vault = vault.NewService(db)
	a.projects = projects.NewService(db, a.vault)
	a.workspaces = workspaces.NewService(db, a.vault)
	a.settings = settings.NewService(db)
	a.ssh = sshconn.NewService(db, a.vault, dataDir)

	logger.Named("app").Info("Velarvo app started")
}

func (a *App) OnDomReady(ctx context.Context) {
	a.ctx = ctx

	accessToken, err := a.keychain.GetAccessToken()
	if err != nil || accessToken == "" {
		return
	}

	refreshToken, err := a.keychain.GetRefreshToken()
	if err != nil || refreshToken == "" {
		return
	}

	a.session.Set(&session.UserSession{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (a *App) OnShutdown(ctx context.Context) {
	a.ctx = ctx

	if a.ssh != nil {
		a.ssh.Shutdown()
	}

	if a.vault != nil {
		a.vault.Lock()
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			logger.Named("app").Errorw("failed to close local storage", "error", err)
		}
	}

	memguard.Purge()

	logger.Named("app").Info("Velarvo app shutting down")
}

func (a *App) OnBeforeClose(ctx context.Context) bool {
	a.session.Clear()
	return false
}

func userDataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	name := "Velarvo"
	if runtime.GOOS == "linux" {
		name = "velarvo"
	}

	return filepath.Join(base, name), nil
}
