package sshconn

import "github.com/Velarvo/velarvo-desktop/internal/storage"

type ListingDataV1 struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword,omitempty"`
	OS          string `json:"os,omitempty"`
}

type SecretDataV1 struct {
	Password string `json:"password,omitempty"`
}

type ListingData = ListingDataV1

type SecretData = SecretDataV1

type ConnectionDTO struct {
	ID          string             `json:"id"`
	WorkspaceID string             `json:"workspaceId"`
	Name        string             `json:"name"`
	Host        string             `json:"host"`
	Port        int                `json:"port"`
	Username    string             `json:"username"`
	HasPassword bool               `json:"hasPassword"`
	OS          string             `json:"os"`
	Connected   bool               `json:"connected"`
	SortOrder   int                `json:"sortOrder"`
	LastUsedAt  *storage.Timestamp `json:"lastUsedAt,omitempty"`
	CreatedAt   storage.Timestamp  `json:"createdAt"`
	UpdatedAt   storage.Timestamp  `json:"updatedAt"`
	Revision    string             `json:"revision"`
}

type CreateConnectionRequest struct {
	WorkspaceID string `json:"workspaceId"`
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type UpdateConnectionRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ClearPassword bool   `json:"clearPassword"`
}

type ConnectionState struct {
	ID        string `json:"id"`
	Connected bool   `json:"connected"`
	OS        string `json:"os"`
}
