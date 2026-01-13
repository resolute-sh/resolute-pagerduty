// Package pagerduty provides PagerDuty integration activities for resolute workflows.
package pagerduty

import (
	"github.com/resolute-sh/resolute/core"
	"go.temporal.io/sdk/worker"
)

const (
	ProviderName    = "resolute-pagerduty"
	ProviderVersion = "1.0.0"
)

// Provider returns the PagerDuty provider for registration.
func Provider() core.Provider {
	return core.NewProvider(ProviderName, ProviderVersion).
		AddActivity("pagerduty.FetchIncidents", FetchIncidentsActivity).
		AddActivity("pagerduty.FetchIncident", FetchIncidentActivity).
		AddActivity("pagerduty.FetchPostmortems", FetchPostmortemsActivity)
}

// RegisterActivities registers all PagerDuty activities with a Temporal worker.
func RegisterActivities(w worker.Worker) {
	core.RegisterProviderActivities(w, Provider())
}
