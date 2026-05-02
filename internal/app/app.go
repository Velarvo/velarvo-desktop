package app

import (
	"context"

	"github.com/Velarvo/velarvo-desktop/internal/client"
	"github.com/Velarvo/velarvo-desktop/internal/keychain"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/session"
)

type App struct {
	ctx      context.Context
	client   *client.Client
	keychain *keychain.Service
	session  *session.Manager
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

func (a *App) OnBeforeClose(ctx context.Context) bool {
	a.session.Clear()
	return false
}
