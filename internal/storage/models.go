package storage

type AppMeta struct {
	Key   string
	Value string
}

type VaultMeta struct {
	SchemaVersion   SchemaVersion
	CryptoVersion   CryptoVersion
	CreatedAt       Timestamp
	KDFID           KDFID
	KDFTime         int
	KDFMemory       int
	KDFThreads      int
	KDFKeyLen       int
	SaltKEK         []byte
	SaltAuth        []byte
	DeviceID        string
	AutoLockSeconds int
}

type DEKEnvelope struct {
	Method       DEKEnvelopeMethod
	Envelope     []byte
	MetadataJSON *string
	CreatedAt    Timestamp
	UpdatedAt    Timestamp
}

type Project struct {
	ID              string
	EncryptedData   []byte
	DataVersion     DataVersion
	SortOrder       int
	CreatedAt       Timestamp
	UpdatedAt       Timestamp
	ServerUpdatedAt *Timestamp
	DeletedAt       *Timestamp
	Revision        string
}

type ProjectIcon struct {
	ProjectID       string
	EncryptedIcon   []byte
	IconMIME        ProjectIconMIME
	LocalStatus     LocalStatus
	SyncState       SyncItemState
	CreatedAt       Timestamp
	UpdatedAt       Timestamp
	ServerUpdatedAt *Timestamp
	Revision        string
}

type SSHConnection struct {
	ID               string
	WorkspaceID      string
	ListingEncrypted []byte
	ListingVersion   DataVersion
	SecretEncrypted  []byte
	SecretVersion    DataVersion
	SortOrder        int
	LastUsedAt       *Timestamp
	CreatedAt        Timestamp
	UpdatedAt        Timestamp
	ServerUpdatedAt  *Timestamp
	DeletedAt        *Timestamp
	Revision         string
}

type Workspace struct {
	ID              string
	ProjectID       string
	EncryptedData   []byte
	DataVersion     DataVersion
	SortOrder       int
	CreatedAt       Timestamp
	UpdatedAt       Timestamp
	ServerUpdatedAt *Timestamp
	DeletedAt       *Timestamp
	Revision        string
}

type Collection struct {
	ID              string
	WorkspaceID     string
	Kind            CollectionKind
	EncryptedData   []byte
	DataVersion     DataVersion
	SortOrder       int
	CreatedAt       Timestamp
	UpdatedAt       Timestamp
	ServerUpdatedAt *Timestamp
	DeletedAt       *Timestamp
	Revision        string
}

type VaultBlock struct {
	ID               string
	WorkspaceID      string
	CollectionID     *string
	Kind             VaultBlockKind
	ListingEncrypted []byte
	ListingVersion   DataVersion
	SecretEncrypted  []byte
	SecretVersion    DataVersion
	LastUsedAt       *Timestamp
	CreatedAt        Timestamp
	UpdatedAt        Timestamp
	ServerUpdatedAt  *Timestamp
	DeletedAt        *Timestamp
	Revision         string
}

type SyncState struct {
	Enabled        bool
	ServerURL      *string
	LastSyncAt     *Timestamp
	LastSyncCursor *string
	Status         SyncStatus
	LastError      *string
	LastErrorAt    *Timestamp
	PendingUploads int
}

type SyncOutboxEntry struct {
	Seq           int64
	EntityType    SyncEntityType
	EntityID      string
	Operation     SyncOperation
	Revision      string
	CreatedAt     Timestamp
	LastAttemptAt *Timestamp
	AttemptCount  int
	LastError     *string
}

type SyncConflict struct {
	ID              string
	EntityType      SyncEntityType
	EntityID        string
	LocalRevision   string
	LocalData       []byte
	LocalUpdatedAt  Timestamp
	ServerRevision  string
	ServerData      []byte
	ServerUpdatedAt Timestamp
	Resolution      *ConflictResolution
	ResolvedAt      *Timestamp
	DetectedAt      Timestamp
}
