// Package api provides a complete GraphQL client for banned.video API
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents the GraphQL API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
}

// NewClient creates a new API client with default settings
func NewClient() *Client {
	return &Client{
		BaseURL: "https://api.banned.video/graphql",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: "banned-cli/1.0",
	}
}

// NewClientWithConfig creates a new API client with custom configuration
func NewClientWithConfig(baseURL, userAgent string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		UserAgent: userAgent,
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents the location of a GraphQL error
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Error implements the error interface for GraphQLError
func (e GraphQLError) Error() string {
	return e.Message
}

// Execute performs a GraphQL query and returns the response
func (c *Client) Execute(ctx context.Context, req *GraphQLRequest) (*GraphQLResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.UserAgent)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	return &gqlResp, nil
}

// ExecuteWithResult performs a GraphQL query and unmarshals the result
func (c *Client) ExecuteWithResult(ctx context.Context, req *GraphQLRequest, result interface{}) error {
	resp, err := c.Execute(ctx, req)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}
