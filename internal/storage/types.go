package storage

import "time"

type Timestamp int64

func Now() Timestamp {
	return Timestamp(time.Now().UTC().UnixMicro())
}

func (t Timestamp) Time() time.Time {
	return time.UnixMicro(int64(t)).UTC()
}

type KDFID int

const (
	KDFArgon2id KDFID = 1
)

func (k KDFID) Valid() bool {
	return k == KDFArgon2id
}

type CryptoVersion int

const (
	CryptoVersionAESGCM  CryptoVersion = 1
	CurrentCryptoVersion               = CryptoVersionAESGCM
)

func (v CryptoVersion) Valid() bool {
	return v >= CryptoVersionAESGCM
}

type SchemaVersion int

const (
	SchemaVersionInitial      SchemaVersion = 1
	CurrentSchemaVersion                    = SchemaVersionInitial
	MaxSupportedSchemaVersion               = CurrentSchemaVersion
)

func (v SchemaVersion) Supported() bool {
	return v >= SchemaVersionInitial && v <= MaxSupportedSchemaVersion
}

type DataVersion int

const (
	DataVersionV1 DataVersion = 1

	CurrentProjectDataVersion       = DataVersionV1
	CurrentWorkspaceDataVersion     = DataVersionV1
	CurrentCollectionDataVersion    = DataVersionV1
	CurrentVaultBlockListingVersion = DataVersionV1
	CurrentVaultBlockSecretVersion  = DataVersionV1
	CurrentSSHListingVersion        = DataVersionV1
	CurrentSSHSecretVersion         = DataVersionV1
)

func (v DataVersion) Valid() bool {
	return v >= DataVersionV1
}

type DEKEnvelopeMethod string

const (
	DEKEnvelopePassword DEKEnvelopeMethod = "password"
)

func (m DEKEnvelopeMethod) Valid() bool {
	switch m {
	case DEKEnvelopePassword:
		return true
	default:
		return false
	}
}

type CollectionKind string

const (
	CollectionKindAPI      CollectionKind = "api"
	CollectionKindSSH      CollectionKind = "ssh"
	CollectionKindDatabase CollectionKind = "database"
	CollectionKindVault    CollectionKind = "vault"
)

func (k CollectionKind) Valid() bool {
	switch k {
	case CollectionKindAPI, CollectionKindSSH, CollectionKindDatabase, CollectionKindVault:
		return true
	default:
		return false
	}
}

type VaultBlockKind string

const (
	VaultBlockCredential    VaultBlockKind = "credential"
	VaultBlockAPIKey        VaultBlockKind = "api_key"
	VaultBlockSSHConnection VaultBlockKind = "ssh_connection"
	VaultBlockDBConnection  VaultBlockKind = "db_connection"
	VaultBlockAPIRequest    VaultBlockKind = "api_request"
	VaultBlockEnvVar        VaultBlockKind = "env_var"
	VaultBlockNote          VaultBlockKind = "note"
	VaultBlockSecret        VaultBlockKind = "secret"
)

func (k VaultBlockKind) Valid() bool {
	switch k {
	case VaultBlockCredential, VaultBlockAPIKey, VaultBlockSSHConnection, VaultBlockDBConnection,
		VaultBlockAPIRequest, VaultBlockEnvVar, VaultBlockNote, VaultBlockSecret:
		return true
	default:
		return false
	}
}

type ProjectIconMIME string

const (
	ProjectIconPNG  ProjectIconMIME = "image/png"
	ProjectIconJPEG ProjectIconMIME = "image/jpeg"
	ProjectIconWebP ProjectIconMIME = "image/webp"
)

func (m ProjectIconMIME) Valid() bool {
	switch m {
	case ProjectIconPNG, ProjectIconJPEG, ProjectIconWebP:
		return true
	default:
		return false
	}
}

type LocalStatus string

const (
	LocalStatusPresent LocalStatus = "present"
	LocalStatusMissing LocalStatus = "missing"
)

func (s LocalStatus) Valid() bool {
	switch s {
	case LocalStatusPresent, LocalStatusMissing:
		return true
	default:
		return false
	}
}

type SyncItemState string

const (
	SyncItemLocalOnly  SyncItemState = "local_only"
	SyncItemSyncing    SyncItemState = "syncing"
	SyncItemSynced     SyncItemState = "synced"
	SyncItemRemoteOnly SyncItemState = "remote_only"
	SyncItemFailed     SyncItemState = "failed"
)

func (s SyncItemState) Valid() bool {
	switch s {
	case SyncItemLocalOnly, SyncItemSyncing, SyncItemSynced, SyncItemRemoteOnly, SyncItemFailed:
		return true
	default:
		return false
	}
}

type SyncStatus string

const (
	SyncStatusIdle    SyncStatus = "idle"
	SyncStatusSyncing SyncStatus = "syncing"
	SyncStatusError   SyncStatus = "error"
	SyncStatusPaused  SyncStatus = "paused"
)

func (s SyncStatus) Valid() bool {
	switch s {
	case SyncStatusIdle, SyncStatusSyncing, SyncStatusError, SyncStatusPaused:
		return true
	default:
		return false
	}
}

type SyncEntityType string

const (
	SyncEntityProject     SyncEntityType = "project"
	SyncEntityProjectIcon SyncEntityType = "project_icon"
	SyncEntityWorkspace   SyncEntityType = "workspace"
	SyncEntityCollection  SyncEntityType = "collection"
	SyncEntityVaultBlock  SyncEntityType = "vault_block"
)

func (t SyncEntityType) Valid() bool {
	switch t {
	case SyncEntityProject, SyncEntityProjectIcon, SyncEntityWorkspace, SyncEntityCollection, SyncEntityVaultBlock:
		return true
	default:
		return false
	}
}

type SyncOperation string

const (
	SyncOperationCreate SyncOperation = "create"
	SyncOperationUpdate SyncOperation = "update"
	SyncOperationDelete SyncOperation = "delete"
)

func (o SyncOperation) Valid() bool {
	switch o {
	case SyncOperationCreate, SyncOperationUpdate, SyncOperationDelete:
		return true
	default:
		return false
	}
}

type ConflictResolution string

const (
	ConflictResolutionLocal  ConflictResolution = "local"
	ConflictResolutionServer ConflictResolution = "server"
	ConflictResolutionMerged ConflictResolution = "merged"
)

func (r ConflictResolution) Valid() bool {
	switch r {
	case ConflictResolutionLocal, ConflictResolutionServer, ConflictResolutionMerged:
		return true
	default:
		return false
	}
}
