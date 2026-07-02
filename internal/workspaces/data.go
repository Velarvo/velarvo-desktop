package workspaces

import "github.com/Velarvo/velarvo-desktop/internal/storage"

type WorkspaceDataV1 struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type WorkspaceData = WorkspaceDataV1

type WorkspaceDTO struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"projectId"`
	Name      string            `json:"name"`
	Color     string            `json:"color"`
	SortOrder int               `json:"sortOrder"`
	CreatedAt storage.Timestamp `json:"createdAt"`
	UpdatedAt storage.Timestamp `json:"updatedAt"`
	Revision  string            `json:"revision"`
}

type CreateWorkspaceRequest struct {
	ProjectID string `json:"projectId"`
	Name      string `json:"name"`
	Color     string `json:"color,omitempty"`
}

type UpdateWorkspaceRequest struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}
