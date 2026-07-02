package app

import (
	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/types"
	"github.com/Velarvo/velarvo-desktop/internal/workspaces"
)

func (a *App) ListWorkspaces(projectID string) types.APIResponse[[]workspaces.WorkspaceDTO] {
	service, err := a.requireWorkspaces()
	if err != nil {
		return errResponse[[]workspaces.WorkspaceDTO](string(apperrors.CodeWorkspacesNotInitialized), "", err)
	}
	items, err := service.ListByProject(a.ctx, projectID)
	if err != nil {
		return errResponse[[]workspaces.WorkspaceDTO](string(apperrors.CodeWorkspaceListFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", items)
}

func (a *App) GetWorkspace(id string) types.APIResponse[workspaces.WorkspaceDTO] {
	service, err := a.requireWorkspaces()
	if err != nil {
		return errResponse[workspaces.WorkspaceDTO](string(apperrors.CodeWorkspacesNotInitialized), "", err)
	}
	item, err := service.Get(a.ctx, id)
	if err != nil {
		return errResponse[workspaces.WorkspaceDTO](string(apperrors.CodeWorkspaceReadFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) CreateWorkspace(req workspaces.CreateWorkspaceRequest) types.APIResponse[workspaces.WorkspaceDTO] {
	service, err := a.requireWorkspaces()
	if err != nil {
		return errResponse[workspaces.WorkspaceDTO](string(apperrors.CodeWorkspacesNotInitialized), "", err)
	}
	item, err := service.Create(a.ctx, req)
	if err != nil {
		return errResponse[workspaces.WorkspaceDTO](string(apperrors.CodeWorkspaceCreateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) UpdateWorkspace(req workspaces.UpdateWorkspaceRequest) types.APIResponse[workspaces.WorkspaceDTO] {
	service, err := a.requireWorkspaces()
	if err != nil {
		return errResponse[workspaces.WorkspaceDTO](string(apperrors.CodeWorkspacesNotInitialized), "", err)
	}
	item, err := service.Update(a.ctx, req)
	if err != nil {
		return errResponse[workspaces.WorkspaceDTO](string(apperrors.CodeWorkspaceUpdateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) DeleteWorkspace(id string) types.APIResponse[map[string]string] {
	service, err := a.requireWorkspaces()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeWorkspacesNotInitialized), "", err)
	}
	if err := service.Delete(a.ctx, id); err != nil {
		return errResponse[map[string]string](string(apperrors.CodeWorkspaceDeleteFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}

func (a *App) requireWorkspaces() (*workspaces.Service, error) {
	if a.workspaces == nil {
		return nil, ErrWorkspacesServiceUnavailable
	}
	return a.workspaces, nil
}
