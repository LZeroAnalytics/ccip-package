package chainlink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Client represents a Chainlink node client
type Client struct {
	NodeID        string
	BaseURL       string
	HTTPClient    *http.Client
	SessionCookie string
	AuthToken     string
}

// ClientManager manages connections to multiple Chainlink nodes
type ClientManager struct {
	clients map[string]*Client
}

// API Response structures
type ChainlinkJob struct {
	ID              string           `json:"id"`
	ExternalJobID   string           `json:"externalJobID"`
	Type            string           `json:"type"`
	SchemaVersion   int              `json:"schemaVersion"`
	Spec            interface{}      `json:"spec"`
	Name            string           `json:"name"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	Errors          []ChainlinkError `json:"errors"`
	MaxTaskDuration string           `json:"maxTaskDuration"`
	PipelineSpec    PipelineSpec     `json:"pipelineSpec"`
}

type ChainlinkError struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Occurrences int       `json:"occurrences"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type PipelineSpec struct {
	ID           int    `json:"id"`
	DotDAGSource string `json:"dotDagSource"`
}

type JobsResponse struct {
	Data []ChainlinkJob `json:"data"`
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

type CreateJobRequest struct {
	TOML string `json:"toml"`
}

type HealthResponse struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type SessionRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type KeysResponse struct {
	Data []Key `json:"data"`
}

type Key struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Type      string `json:"type"`
}

// NewClientManager creates a new client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*Client),
	}
}

// GetClient returns a client for a specific node
func (m *ClientManager) GetClient(nodeID string) *Client {
	return m.clients[nodeID]
}

// AddNode adds a new node to the client manager
func (m *ClientManager) AddNode(nodeID, url, email, password string) error {
	client := &Client{
		NodeID:  nodeID,
		BaseURL: strings.TrimSuffix(url, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		SessionCookie: email,
		AuthToken:     password,
	}

	// Attempt to authenticate
	if err := client.authenticate(context.Background()); err != nil {
		return fmt.Errorf("failed to authenticate with node %s: %w", nodeID, err)
	}

	m.clients[nodeID] = client
	return nil
}

// =============================================================================
// AUTHENTICATION METHODS
// =============================================================================

// authenticate performs authentication with the Chainlink node
func (c *Client) authenticate(ctx context.Context) error {
	// Try session-based auth first (newer Chainlink versions)
	if err := c.authenticateSession(ctx); err != nil {
		log.Printf("Session auth failed for %s, trying basic auth: %v", c.NodeID, err)
		// Fall back to basic auth for older versions
		return c.authenticateBasic(ctx)
	}
	return nil
}

// authenticateSession uses session-based authentication
func (c *Client) authenticateSession(ctx context.Context) error {
	sessionReq := SessionRequest{
		Email:    c.SessionCookie,
		Password: c.AuthToken,
	}

	reqBody, _ := json.Marshal(sessionReq)
	resp, err := c.doRequest(ctx, "POST", "/sessions", bytes.NewReader(reqBody), nil)
	if err != nil {
		return fmt.Errorf("session request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("session auth failed: %d - %s", resp.StatusCode, string(body))
	}

	// Extract session ID from cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "clsession" {
			c.SessionCookie = cookie.Value
			return nil
		}
	}

	return fmt.Errorf("no session cookie found")
}

// authenticateBasic uses basic authentication (older method)
func (c *Client) authenticateBasic(ctx context.Context) error {
	loginReq := LoginRequest{
		Email:    c.SessionCookie,
		Password: c.AuthToken,
	}

	reqBody, _ := json.Marshal(loginReq)
	resp, err := c.doRequest(ctx, "POST", "/v2/authenticate", bytes.NewReader(reqBody), nil)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("basic auth failed: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// =============================================================================
// JOB MANAGEMENT METHODS
// =============================================================================

// CreateJob creates a job on the Chainlink node
func (c *Client) CreateJob(ctx context.Context, spec string) (*ChainlinkJob, error) {
	log.Printf("📝 Creating job on node %s", c.NodeID)

	// Create job request with TOML spec
	createReq := CreateJobRequest{
		TOML: spec,
	}

	reqBody, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create job request: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := c.doRequest(ctx, "POST", "/v2/jobs", bytes.NewReader(reqBody), headers)
	if err != nil {
		return nil, fmt.Errorf("create job request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create job failed: %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data ChainlinkJob `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		// Some versions return the job directly
		var job ChainlinkJob
		if err2 := json.Unmarshal(body, &job); err2 != nil {
			return nil, fmt.Errorf("failed to parse job response: %w", err)
		}
		response.Data = job
	}

	log.Printf("✅ Job created successfully on node %s: %s", c.NodeID, response.Data.ID)
	return &response.Data, nil
}

// GetJob retrieves a specific job from the Chainlink node
func (c *Client) GetJob(ctx context.Context, jobID string) (*ChainlinkJob, error) {
	log.Printf("🔍 Getting job %s from node %s", jobID, c.NodeID)

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/v2/jobs/%s", jobID), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get job request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get job failed: %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data ChainlinkJob `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}

	return &response.Data, nil
}

// ListJobs lists all jobs on the Chainlink node
func (c *Client) ListJobs(ctx context.Context) ([]ChainlinkJob, error) {
	log.Printf("📋 Listing jobs from node %s", c.NodeID)

	resp, err := c.doRequest(ctx, "GET", "/v2/jobs", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list jobs request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list jobs failed: %d - %s", resp.StatusCode, string(body))
	}

	var response JobsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}

	log.Printf("📋 Found %d jobs on node %s", len(response.Data), c.NodeID)
	return response.Data, nil
}

// DeleteJob deletes a job from the Chainlink node
func (c *Client) DeleteJob(ctx context.Context, jobID string) error {
	log.Printf("🗑️ Deleting job %s from node %s", jobID, c.NodeID)

	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/v2/jobs/%s", jobID), nil, nil)
	if err != nil {
		return fmt.Errorf("delete job request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("job %s not found", jobID)
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete job failed: %d - %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ Job %s deleted successfully from node %s", jobID, c.NodeID)
	return nil
}

// =============================================================================
// KEY MANAGEMENT METHODS
// =============================================================================

// GetETHKeys returns all ETH keys from the node
func (c *Client) GetETHKeys(ctx context.Context) ([]Key, error) {
	return c.getKeys(ctx, "/v2/keys/eth")
}

// GetP2PKeys returns all P2P keys from the node
func (c *Client) GetP2PKeys(ctx context.Context) ([]Key, error) {
	return c.getKeys(ctx, "/v2/keys/p2p")
}

// GetOCRKeys returns all OCR keys from the node
func (c *Client) GetOCRKeys(ctx context.Context) ([]Key, error) {
	return c.getKeys(ctx, "/v2/keys/ocr")
}

// GetCSAKeys returns all CSA keys from the node
func (c *Client) GetCSAKeys(ctx context.Context) ([]Key, error) {
	return c.getKeys(ctx, "/v2/keys/csa")
}

// getKeys is a generic method to fetch keys of any type
func (c *Client) getKeys(ctx context.Context, endpoint string) ([]Key, error) {
	resp, err := c.doRequest(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("get keys request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get keys failed: %d - %s", resp.StatusCode, string(body))
	}

	var response KeysResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse keys response: %w", err)
	}

	return response.Data, nil
}

// =============================================================================
// HEALTH AND STATUS METHODS
// =============================================================================

// IsHealthy checks if the Chainlink node is healthy
func (c *Client) IsHealthy(ctx context.Context) bool {
	resp, err := c.doRequest(ctx, "GET", "/health", nil, nil)
	if err != nil {
		log.Printf("⚠️ Health check failed for node %s: %v", c.NodeID, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Node %s unhealthy: %d", c.NodeID, resp.StatusCode)
		return false
	}

	var health HealthResponse
	if body, err := io.ReadAll(resp.Body); err == nil {
		if json.Unmarshal(body, &health) == nil {
			log.Printf("✅ Node %s healthy (version: %s)", c.NodeID, health.Version)
			return health.Healthy
		}
	}

	log.Printf("✅ Node %s healthy", c.NodeID)
	return true
}

// GetVersion returns the Chainlink node version
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	resp, err := c.doRequest(ctx, "GET", "/v2/build_info", nil, nil)
	if err != nil {
		return "", fmt.Errorf("version request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get version failed: %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse version response: %w", err)
	}

	return response.Data.Version, nil
}

// =============================================================================
// LOW-LEVEL HTTP CLIENT METHODS
// =============================================================================

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", "Job-Distributor/1.0")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add authentication
	if c.SessionCookie != "" {
		// Session-based auth
		req.AddCookie(&http.Cookie{
			Name:  "clsession",
			Value: c.SessionCookie,
		})
	} else {
		// Basic auth
		req.SetBasicAuth(c.SessionCookie, c.AuthToken)
	}

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("🔐 Authentication failed for node %s, attempting re-auth", c.NodeID)
		resp.Body.Close()

		// Try to re-authenticate
		if authErr := c.authenticate(ctx); authErr != nil {
			return nil, fmt.Errorf("re-authentication failed: %w", authErr)
		}

		// Retry the original request
		return c.doRequest(ctx, method, path, body, headers)
	}

	return resp, nil
}

// =============================================================================
// UTILITY METHODS
// =============================================================================

// ExtractJobIDFromTOML extracts the externalJobID from a TOML job spec
func ExtractJobIDFromTOML(tomlSpec string) string {
	lines := strings.Split(tomlSpec, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "externalJobID") {
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			}
		}
	}
	return ""
}

// ValidateJobSpec validates a TOML job specification
func ValidateJobSpec(tomlSpec string) error {
	if tomlSpec == "" {
		return fmt.Errorf("job spec cannot be empty")
	}

	requiredFields := []string{"type", "schemaVersion"}
	for _, field := range requiredFields {
		if !strings.Contains(tomlSpec, field) {
			return fmt.Errorf("job spec missing required field: %s", field)
		}
	}

	return nil
}
