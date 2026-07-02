package app

import (
	"github.com/Velarvo/velarvo-desktop/internal/apperrors"
	"github.com/Velarvo/velarvo-desktop/internal/projects"
	"github.com/Velarvo/velarvo-desktop/internal/types"
)

func (a *App) ListProjects() types.APIResponse[[]projects.ProjectDTO] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[[]projects.ProjectDTO](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	items, err := service.List(a.ctx)
	if err != nil {
		return errResponse[[]projects.ProjectDTO](string(apperrors.CodeProjectListFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", items)
}

func (a *App) GetProject(id string) types.APIResponse[projects.ProjectDTO] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[projects.ProjectDTO](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	item, err := service.Get(a.ctx, id)
	if err != nil {
		return errResponse[projects.ProjectDTO](string(apperrors.CodeProjectReadFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) CreateProject(req projects.CreateProjectRequest) types.APIResponse[projects.ProjectDTO] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[projects.ProjectDTO](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	item, err := service.Create(a.ctx, req)
	if err != nil {
		return errResponse[projects.ProjectDTO](string(apperrors.CodeProjectCreateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) UpdateProject(req projects.UpdateProjectRequest) types.APIResponse[projects.ProjectDTO] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[projects.ProjectDTO](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	item, err := service.Update(a.ctx, req)
	if err != nil {
		return errResponse[projects.ProjectDTO](string(apperrors.CodeProjectUpdateFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) DeleteProject(id string) types.APIResponse[map[string]string] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	if err := service.Delete(a.ctx, id); err != nil {
		return errResponse[map[string]string](string(apperrors.CodeProjectDeleteFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}

func (a *App) SetProjectIcon(req projects.SetProjectIconRequest) types.APIResponse[projects.ProjectIconDTO] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[projects.ProjectIconDTO](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	item, err := service.SetIcon(a.ctx, req)
	if err != nil {
		return errResponse[projects.ProjectIconDTO](string(apperrors.CodeProjectIconSetFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) GetProjectIcon(projectID string) types.APIResponse[projects.ProjectIconDTO] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[projects.ProjectIconDTO](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	item, err := service.GetIcon(a.ctx, projectID)
	if err != nil {
		return errResponse[projects.ProjectIconDTO](string(apperrors.CodeProjectIconReadFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", *item)
}

func (a *App) DeleteProjectIcon(projectID string) types.APIResponse[map[string]string] {
	service, err := a.requireProjects()
	if err != nil {
		return errResponse[map[string]string](string(apperrors.CodeProjectsNotInitialized), "", err)
	}
	if err := service.DeleteIcon(a.ctx, projectID); err != nil {
		return errResponse[map[string]string](string(apperrors.CodeProjectIconDeleteFailed), "", err)
	}
	return successResponse(string(apperrors.CodeOK), "", map[string]string{})
}

func (a *App) requireProjects() (*projects.Service, error) {
	if a.projects == nil {
		return nil, ErrProjectsServiceUnavailable
	}
	return a.projects, nil
}
