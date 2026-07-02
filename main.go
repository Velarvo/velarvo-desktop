package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/Velarvo/velarvo-desktop/internal/app"
	"github.com/Velarvo/velarvo-desktop/internal/client"
	"github.com/Velarvo/velarvo-desktop/internal/configs"
	"github.com/Velarvo/velarvo-desktop/internal/doubleclick"
	"github.com/Velarvo/velarvo-desktop/internal/keychain"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg := configs.LoadConfig()
	logger.MustInit(logger.Options{
		Level:       cfg.LogLevel,
		Development: cfg.DevMode,
	})
	defer func() {
		_ = logger.Close()
	}()

	log := logger.Named("bootstrap")
	log.Infow("starting Velarvo Desktop", "baseURL", cfg.BaseURL, "devMode", cfg.DevMode)

	doubleclick.Enable()

	kc := keychain.New()
	c := client.New(cfg.BaseURL, kc)
	a := app.New(c, kc)
	appMenu := menu.NewMenu()

	err := wails.Run(&options.App{
		Title:            "Velarvo Desktop",
		Width:            1024,
		Height:           768,
		MinWidth:         800,
		Menu:             appMenu,
		MinHeight:        600,
		WindowStartState: options.Maximised,
		Frameless:        false,
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			Appearance:           mac.NSAppearanceNameDarkAqua,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			a.Startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			a.OnShutdown(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			a.OnDomReady(ctx)
			setupTrafficLights()
		},
		Bind: []interface{}{a},
	})

	if err != nil {
		log.Fatalw("app failed to start", "error", err)
	}
}
