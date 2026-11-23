package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/Velarvo/velarvo-desktop/internal/apperr"
	"github.com/Velarvo/velarvo-desktop/internal/db"
	"github.com/Velarvo/velarvo-desktop/internal/files"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/projects"
	"github.com/Velarvo/velarvo-desktop/internal/settings"
	"github.com/Velarvo/velarvo-desktop/internal/sshconn"
	"github.com/Velarvo/velarvo-desktop/internal/vault"
	"github.com/Velarvo/velarvo-desktop/internal/workspaces"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const appName = "Velarvo"

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(); err != nil {
		slog.Error("velarvo exited with an error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	level := os.Getenv("VELARVO_LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	if err := logger.Setup(level); err != nil {
		return fmt.Errorf("invalid VELARVO_LOG_LEVEL: %w", err)
	}

	dataDir, err := dataDirectory()
	if err != nil {
		return err
	}
	database, err := db.Open(filepath.Join(dataDir, "vault.db"))
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	dist, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		return err
	}

	keys := &vault.Vault{}
	clients := sshconn.NewClients(filepath.Join(dataDir, "known_hosts"))

	app := application.New(application.Options{
		Name:        appName,
		Description: "Developer workspace for SSH, SFTP, databases and APIs",
		LogLevel:    slog.LevelWarn,

		Services: []application.Service{
			application.NewService(vault.NewService(database, keys)),
			application.NewService(settings.NewService(database)),
			application.NewService(projects.NewService(database, keys)),
			application.NewService(workspaces.NewService(database, keys)),
			application.NewService(sshconn.NewService(database, keys, clients)),
			application.NewService(sshconn.NewTerminal(clients)),
			application.NewService(sshconn.NewSFTP(clients)),
			application.NewService(&files.Service{}),
			application.NewService(&logger.Frontend{}),
		},
		MarshalError: apperr.Marshal,
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(dist),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	menu := application.NewMenu()
	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu)
		menu.AddRole(application.EditMenu)
		menu.AddRole(application.WindowMenu)
	}
	app.Menu.Set(menu)

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            appName,
		Width:            1024,
		Height:           768,
		MinWidth:         800,
		MinHeight:        600,
		StartState:       application.WindowStateMaximised,
		BackgroundColour: application.NewRGB(13, 13, 13),
		URL:              "/",
		Mac: application.MacWindow{
			TitleBar:   application.MacTitleBarHiddenInsetUnified,
			Backdrop:   application.MacBackdropTranslucent,
			Appearance: application.NSAppearanceNameDarkAqua,
		},
	})

	var maximise sync.Once
	stopMaximise := window.OnWindowEvent(events.Mac.WindowDidBecomeKey, func(*application.WindowEvent) {
		maximise.Do(func() { window.Maximise() })
	})
	defer stopMaximise()

	slog.Info("velarvo started", "dataDir", dataDir)
	return app.Run()
}

func dataDirectory() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	name := appName
	if runtime.GOOS == "linux" {
		name = "velarvo"
	}
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create data directory: %w", err)
	}
	return dir, nil
}
