package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Velarvo/velarvo-desktop/internal/keychain"
	"github.com/Velarvo/velarvo-desktop/internal/types"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	keychain   *keychain.Service
}

func New(baseURL string, kc *keychain.Service) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		keychain:   kc,
	}
}

func Post[T any](c *Client, endpoint string, payload any) (*types.APIResponse[T], error) {
	return do[T](c, "POST", endpoint, payload, false)
}

func AuthGet[T any](c *Client, endpoint string) (*types.APIResponse[T], error) {
	return do[T](c, "GET", endpoint, nil, true)
}

func AuthPost[T any](c *Client, endpoint string, payload any) (*types.APIResponse[T], error) {
	return do[T](c, "POST", endpoint, payload, true)
}

func AuthPatch[T any](c *Client, endpoint string, payload any) (*types.APIResponse[T], error) {
	return do[T](c, "PATCH", endpoint, payload, true)
}

func AuthDelete[T any](c *Client, endpoint string, payload any) (*types.APIResponse[T], error) {
	return do[T](c, "DELETE", endpoint, payload, true)
}

func do[T any](c *Client, method, endpoint string, payload any, authenticated bool) (*types.APIResponse[T], error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if authenticated {
		token, err := c.keychain.GetAccessToken()
		if err != nil {
			return nil, fmt.Errorf("get access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp types.APIResponse[T]
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &apiResp, nil
}
