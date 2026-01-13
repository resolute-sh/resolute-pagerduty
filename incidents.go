package pagerduty

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/resolute-sh/resolute/core"
	transform "github.com/resolute-sh/resolute-transform"
)

// FetchIncidentsInput is the input for FetchIncidentsActivity.
type FetchIncidentsInput struct {
	APIKey string
	Since  *time.Time
	Until  *time.Time
	Limit  int
}

// FetchIncidentsOutput is the output of FetchIncidentsActivity.
type FetchIncidentsOutput struct {
	Ref   core.DataRef
	Count int
	Total int
}

// FetchIncidentsActivity fetches incidents from PagerDuty and stores them.
func FetchIncidentsActivity(ctx context.Context, input FetchIncidentsInput) (FetchIncidentsOutput, error) {
	client := NewClient(ClientConfig{
		APIKey: input.APIKey,
	})

	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}

	result, err := client.ListIncidents(ctx, input.Since, input.Until, limit)
	if err != nil {
		return FetchIncidentsOutput{}, fmt.Errorf("list incidents: %w", err)
	}

	docs := make([]transform.Document, 0, len(result.Incidents))
	for _, incident := range result.Incidents {
		doc := incidentToDocument(incident)
		docs = append(docs, doc)
	}

	ref, err := transform.StoreDocuments(ctx, docs)
	if err != nil {
		return FetchIncidentsOutput{}, fmt.Errorf("store documents: %w", err)
	}

	return FetchIncidentsOutput{
		Ref:   ref,
		Count: len(docs),
		Total: result.Total,
	}, nil
}

// FetchIncidentInput is the input for FetchIncidentActivity.
type FetchIncidentInput struct {
	APIKey     string
	IncidentID string
}

// FetchIncidentOutput is the output of FetchIncidentActivity.
type FetchIncidentOutput struct {
	Document transform.Document
	Found    bool
}

// FetchIncidentActivity fetches a single incident by ID.
func FetchIncidentActivity(ctx context.Context, input FetchIncidentInput) (FetchIncidentOutput, error) {
	client := NewClient(ClientConfig{
		APIKey: input.APIKey,
	})

	incident, err := client.GetIncident(ctx, input.IncidentID)
	if err != nil {
		return FetchIncidentOutput{}, fmt.Errorf("get incident: %w", err)
	}

	return FetchIncidentOutput{
		Document: incidentToDocument(*incident),
		Found:    true,
	}, nil
}

// FetchPostmortemsInput is the input for FetchPostmortemsActivity.
type FetchPostmortemsInput struct {
	APIKey string
	Since  *time.Time
	Limit  int
}

// FetchPostmortemsOutput is the output of FetchPostmortemsActivity.
type FetchPostmortemsOutput struct {
	Ref   core.DataRef
	Count int
}

// FetchPostmortemsActivity fetches postmortems from PagerDuty and stores them.
func FetchPostmortemsActivity(ctx context.Context, input FetchPostmortemsInput) (FetchPostmortemsOutput, error) {
	client := NewClient(ClientConfig{
		APIKey: input.APIKey,
	})

	limit := input.Limit
	if limit <= 0 {
		limit = 100
	}

	result, err := client.ListIncidents(ctx, input.Since, nil, limit)
	if err != nil {
		return FetchPostmortemsOutput{}, fmt.Errorf("list incidents: %w", err)
	}

	docs := make([]transform.Document, 0)
	for _, incident := range result.Incidents {
		if incident.Status == "resolved" {
			doc := incidentToDocument(incident)
			doc.Metadata["document_type"] = "postmortem"
			docs = append(docs, doc)
		}
	}

	ref, err := transform.StoreDocuments(ctx, docs)
	if err != nil {
		return FetchPostmortemsOutput{}, fmt.Errorf("store documents: %w", err)
	}

	return FetchPostmortemsOutput{
		Ref:   ref,
		Count: len(docs),
	}, nil
}

func incidentToDocument(incident Incident) transform.Document {
	var contentParts []string
	contentParts = append(contentParts, incident.Summary)

	if incident.Description != "" {
		contentParts = append(contentParts, incident.Description)
	}

	content := strings.Join(contentParts, "\n\n")

	metadata := map[string]string{
		"incident_id": incident.ID,
		"status":      incident.Status,
		"urgency":     incident.Urgency,
		"service":     incident.Service.Name,
	}

	if incident.Priority != nil {
		metadata["priority"] = incident.Priority.Name
	}

	if len(incident.Assignments) > 0 {
		metadata["assignee"] = incident.Assignments[0].Assignee.Name
	}

	return transform.Document{
		ID:        incident.ID,
		Content:   content,
		Title:     incident.Summary,
		Source:    "pagerduty",
		URL:       incident.HTMLURL,
		Metadata:  metadata,
		UpdatedAt: incident.UpdatedAt,
	}
}

// FetchIncidents creates a node for fetching PagerDuty incidents.
func FetchIncidents(input FetchIncidentsInput) *core.Node[FetchIncidentsInput, FetchIncidentsOutput] {
	return core.NewNode("pagerduty.FetchIncidents", FetchIncidentsActivity, input)
}

// FetchIncident creates a node for fetching a single PagerDuty incident.
func FetchIncident(input FetchIncidentInput) *core.Node[FetchIncidentInput, FetchIncidentOutput] {
	return core.NewNode("pagerduty.FetchIncident", FetchIncidentActivity, input)
}

// FetchPostmortems creates a node for fetching PagerDuty postmortems.
func FetchPostmortems(input FetchPostmortemsInput) *core.Node[FetchPostmortemsInput, FetchPostmortemsOutput] {
	return core.NewNode("pagerduty.FetchPostmortems", FetchPostmortemsActivity, input)
}
