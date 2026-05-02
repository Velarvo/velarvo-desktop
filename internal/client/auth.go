package client

import "github.com/Velarvo/velarvo-desktop/internal/types"

func (c *Client) Register(payload *types.RegisterPayload) (*types.APIResponse[types.RegisterData], error) {
	return Post[types.RegisterData](c, "/auth/register", payload)
}

func (c *Client) Login(payload *types.LoginPayload) (*types.APIResponse[types.LoginData], error) {
	return Post[types.LoginData](c, "/auth/login", payload)
}

func (c *Client) Refresh(payload *types.RefreshPayload) (*types.APIResponse[types.TokenPair], error) {
	return Post[types.TokenPair](c, "/auth/refresh", payload)
}

func (c *Client) Logout(payload *types.RefreshPayload) (*types.APIResponse[any], error) {
	return Post[any](c, "/auth/logout", payload)
}

func (c *Client) GetCurrentUser() (*types.APIResponse[types.CurrentUserData], error) {
	return AuthGet[types.CurrentUserData](c, "/auth/me")
}
