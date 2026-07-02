package app

import (
	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/sshconn"
	"github.com/Velarvo/velarvo-desktop/internal/types"
)

func (a *App) ListSSHConnections(workspaceID string) types.APIResponse[[]sshconn.ConnectionDTO] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[[]sshconn.ConnectionDTO](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	items, err := service.List(a.ctx, workspaceID)
	if err != nil {
		return errResponse[[]sshconn.ConnectionDTO](string(apperrors.CodeSSHListFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", items)
}

func (a *App) GetSSHConnection(id string) types.APIResponse[sshconn.ConnectionDTO] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[sshconn.ConnectionDTO](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	item, err := service.Get(a.ctx, id)
	if err != nil {
		return errResponse[sshconn.ConnectionDTO](string(apperrors.CodeSSHReadFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) CreateSSHConnection(req sshconn.CreateConnectionRequest) types.APIResponse[sshconn.ConnectionDTO] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[sshconn.ConnectionDTO](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	item, err := service.Create(a.ctx, req)
	if err != nil {
		return errResponse[sshconn.ConnectionDTO](string(apperrors.CodeSSHCreateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) UpdateSSHConnection(req sshconn.UpdateConnectionRequest) types.APIResponse[sshconn.ConnectionDTO] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[sshconn.ConnectionDTO](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	item, err := service.Update(a.ctx, req)
	if err != nil {
		return errResponse[sshconn.ConnectionDTO](string(apperrors.CodeSSHUpdateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) DeleteSSHConnection(id string) types.APIResponse[map[string]string] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	if err := service.Delete(a.ctx, id); err != nil {
		return errResponse[map[string]string](string(apperrors.CodeSSHDeleteFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}

func (a *App) ConnectSSHConnection(id string) types.APIResponse[sshconn.ConnectionState] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[sshconn.ConnectionState](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	state, err := service.Connect(a.ctx, id)
	if err != nil {
		return errResponse[sshconn.ConnectionState](string(apperrors.CodeSSHConnectFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *state)
}

func (a *App) DisconnectSSHConnection(id string) types.APIResponse[sshconn.ConnectionState] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[sshconn.ConnectionState](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	state, err := service.Disconnect(a.ctx, id)
	if err != nil {
		return errResponse[sshconn.ConnectionState](string(apperrors.CodeSSHDisconnectFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *state)
}

func (a *App) ListConnectedSSH() types.APIResponse[[]string] {
	service, err := a.requireSSH()
	if err != nil {
		return errResponse[[]string](string(apperrors.CodeSSHNotInitialized), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", service.ConnectedIDs())
}

func (a *App) requireSSH() (*sshconn.Service, error) {
	if a.ssh == nil {
		return nil, ErrSSHServiceUnavailable
	}
	return a.ssh, nil
}
