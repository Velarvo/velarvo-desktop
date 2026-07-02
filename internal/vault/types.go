package vault

type SetupRequest struct {
	MasterPassword string `json:"masterPassword"`
}

type UnlockRequest struct {
	MasterPassword string `json:"masterPassword"`
}

type VaultState struct {
	IsSetup         bool `json:"isSetup"`
	IsUnlocked      bool `json:"isUnlocked"`
	AutoLockSeconds int  `json:"autoLockSeconds"`
}
