package types

type RegisterPayload struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
}

type LoginPayload struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	DeviceClientId string `json:"deviceClientId"`
	DeviceOs       string `json:"deviceOs"`
	DeviceName     string `json:"deviceName"`
}

type RefreshPayload struct {
	RefreshToken string `json:"refreshToken"`
}

type RegisterData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type LoginData struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	User         UserData   `json:"user"`
	Device       DeviceData `json:"device"`
}

type UserData struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

type DeviceData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type CurrentUserData struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Premium   bool    `json:"premium"`
}
