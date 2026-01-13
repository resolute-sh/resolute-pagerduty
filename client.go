package pagerduty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.pagerduty.com"

// Client is a PagerDuty REST API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// ClientConfig contains configuration for creating a PagerDuty client.
type ClientConfig struct {
	APIKey  string
	Timeout time.Duration
}

// NewClient creates a new PagerDuty client.
func NewClient(cfg ClientConfig) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Incident represents a PagerDuty incident.
type Incident struct {
	ID               string           `json:"id"`
	Type             string           `json:"type"`
	Summary          string           `json:"summary"`
	Description      string           `json:"description"`
	Status           string           `json:"status"`
	Urgency          string           `json:"urgency"`
	Priority         *Priority        `json:"priority"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	ResolvedAt       *time.Time       `json:"resolved_at"`
	Service          Service          `json:"service"`
	Assignments      []Assignment     `json:"assignments"`
	EscalationPolicy EscalationPolicy `json:"escalation_policy"`
	HTMLURL          string           `json:"html_url"`
}

// Priority represents incident priority.
type Priority struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// Service represents a PagerDuty service.
type Service struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// Assignment represents an incident assignment.
type Assignment struct {
	At       time.Time `json:"at"`
	Assignee Assignee  `json:"assignee"`
}

// Assignee represents an assigned user.
type Assignee struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// EscalationPolicy represents an escalation policy.
type EscalationPolicy struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// Postmortem represents a PagerDuty postmortem.
type Postmortem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	HTMLURL     string    `json:"html_url"`
}

// IncidentListResponse represents the response from listing incidents.
type IncidentListResponse struct {
	Incidents []Incident `json:"incidents"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
	Total     int        `json:"total"`
	More      bool       `json:"more"`
}

// ListIncidents fetches incidents.
func (c *Client) ListIncidents(ctx context.Context, since *time.Time, until *time.Time, limit int) (*IncidentListResponse, error) {
	if limit <= 0 {
		limit = 25
	}

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))

	if since != nil {
		params.Set("since", since.Format(time.RFC3339))
	}
	if until != nil {
		params.Set("until", until.Format(time.RFC3339))
	}

	endpoint := fmt.Sprintf("%s/incidents?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pagerduty API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result IncidentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GetIncident fetches a single incident by ID.
func (c *Client) GetIncident(ctx context.Context, incidentID string) (*Incident, error) {
	endpoint := fmt.Sprintf("%s/incidents/%s", baseURL, incidentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pagerduty API error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Incident Incident `json:"incident"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result.Incident, nil
}

func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Token token="+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}
