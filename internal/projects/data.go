package projects

import "github.com/Velarvo/velarvo-desktop/internal/storage"

type ProjectDataV1 struct {
	Name     string          `json:"name"`
	Color    string          `json:"color,omitempty"`
	Settings ProjectSettings `json:"settings,omitempty"`
}

type ProjectSettings struct {
	DefaultWorkspaceID string `json:"defaultWorkspaceId,omitempty"`
}

type ProjectData = ProjectDataV1

type ProjectDTO struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Color     string            `json:"color"`
	SortOrder int               `json:"sortOrder"`
	HasIcon   bool              `json:"hasIcon"`
	CreatedAt storage.Timestamp `json:"createdAt"`
	UpdatedAt storage.Timestamp `json:"updatedAt"`
	Revision  string            `json:"revision"`
}

type CreateProjectRequest struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type UpdateProjectRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type ProjectIconDTO struct {
	ProjectID  string                  `json:"projectId"`
	Status     storage.LocalStatus     `json:"status"`
	SyncState  storage.SyncItemState   `json:"syncState"`
	MIME       storage.ProjectIconMIME `json:"mime,omitempty"`
	DataBase64 string                  `json:"dataBase64,omitempty"`
	UpdatedAt  storage.Timestamp       `json:"updatedAt,omitempty"`
}

type SetProjectIconRequest struct {
	ProjectID string                  `json:"projectId"`
	MIME      storage.ProjectIconMIME `json:"mime"`
	Data      []byte                  `json:"data"`
}
