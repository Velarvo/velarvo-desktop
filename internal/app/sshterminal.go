package app

import (
	"encoding/base64"

	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/sshconn"
	"github.com/Velarvo/velarvo-desktop/internal/types"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func sshTermDataEvent(sessionID string) string { return "ssh:term:data:" + sessionID }
func sshTermExitEvent(sessionID string) string { return "ssh:term:exit:" + sessionID }

func (a *App) OpenSSHTerminal(connectionID string, cols int, rows int) types.APIResponse[string] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[string](string(apperrors.CodeSSHNotInitialized), "", err)
	}

	ctx := a.ctx
	callbacks := sshconn.TerminalCallbacks{
		OnData: func(sessionID string, chunk []byte) {
			wailsruntime.EventsEmit(
				ctx,
				sshTermDataEvent(sessionID),
				base64.StdEncoding.EncodeToString(chunk),
			)
		},
		OnExit: func(sessionID string) {
			wailsruntime.EventsEmit(ctx, sshTermExitEvent(sessionID))
		},
	}

	sessionID, err := service.OpenTerminal(connectionID, cols, rows, callbacks)
	if err != nil {
		return errResponse[string](string(apperrors.CodeSSHTerminalFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", sessionID)
}

func (a *App) WriteSSHTerminal(sessionID string, data string) types.APIResponse[map[string]string] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	if err := service.WriteTerminal(sessionID, []byte(data)); err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHTerminalFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}

func (a *App) ResizeSSHTerminal(sessionID string, cols int, rows int) types.APIResponse[map[string]string] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	if err := service.ResizeTerminal(sessionID, cols, rows); err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHTerminalFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}

func (a *App) CloseSSHTerminal(sessionID string) types.APIResponse[map[string]string] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	service.CloseTerminal(sessionID)
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}
