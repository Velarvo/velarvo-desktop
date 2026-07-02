package types

type APIResponse[T any] struct {
	Success    bool              `json:"success"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Params     map[string]string `json:"params,omitempty"`
	Data       *T                `json:"data,omitempty"`
	Error      string            `json:"error,omitempty"`
	StatusCode int               `json:"statusCode,omitempty"`
}
