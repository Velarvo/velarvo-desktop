package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Velarvo/velarvo-desktop/internal/keychain"
	"github.com/Velarvo/velarvo-desktop/internal/logger"
	"github.com/Velarvo/velarvo-desktop/internal/types"
	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	keychain   *keychain.Service
}

func clientLog() *zap.SugaredLogger {
	return logger.Named("client")
}

func New(baseURL string, kc *keychain.Service) *Client {
	clientLog().Infow("initializing API client", "baseURL", baseURL)

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
	log := clientLog()

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Errorw("failed to marshal request payload", "method", method, "endpoint", endpoint, "error", err)
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	log.Debugw("sending API request", "method", method, "endpoint", endpoint, "authenticated", authenticated)

	req, err := http.NewRequestWithContext(context.Background(), method, c.baseURL+endpoint, body)
	if err != nil {
		log.Errorw("failed to create API request", "method", method, "endpoint", endpoint, "error", err)
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if authenticated {
		token, err := c.keychain.GetAccessToken()
		if err != nil {
			log.Errorw("failed to load access token", "method", method, "endpoint", endpoint, "error", err)
			return nil, fmt.Errorf("get access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Errorw("API request failed", "method", method, "endpoint", endpoint, "error", err)
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorw("failed to read API response body", "method", method, "endpoint", endpoint, "statusCode", resp.StatusCode, "error", err)
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp types.APIResponse[T]
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		log.Errorw("failed to decode API response", "method", method, "endpoint", endpoint, "statusCode", resp.StatusCode, "error", err)
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if apiResp.Success {
		log.Debugw("API request completed", "method", method, "endpoint", endpoint, "statusCode", resp.StatusCode, "code", apiResp.Code)
	} else {
		log.Warnw("API request returned unsuccessful response", "method", method, "endpoint", endpoint, "statusCode", resp.StatusCode, "code", apiResp.Code, "message", apiResp.Message)
	}

	return &apiResp, nil
}
