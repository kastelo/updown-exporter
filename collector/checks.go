package collector

import (
	"cmp"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"

	"kastelo.dev/updown-exporter/updown"
)

// ChecksCollector is a type that represents updown Checks
type ChecksCollector struct {
	System  System
	Client  *updown.Client
	Log     logr.Logger
	Enabled *prometheus.Desc
}

// NewChecksCollector is a function that returns a new ChecksCollector
func NewChecksCollector(s System, client *updown.Client, log logr.Logger) *ChecksCollector {
	subsystem := "checks"
	return &ChecksCollector{
		System: s,
		Client: client,
		Log:    log,
		Enabled: prometheus.NewDesc(
			prometheus.BuildFQName(s.Namespace, subsystem, "up"),
			"status of check",
			[]string{
				"url",
				"alias",
			},
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *ChecksCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.Log.WithName("Collect")

	checks, err := c.Client.GetChecks()
	if err != nil {
		log.Info("Unable to get Checks")
		return
	}

	for _, check := range checks {
		ch <- prometheus.MustNewConstMetric(
			c.Enabled,
			prometheus.CounterValue,
			boolFloat(!check.Down),
			check.URL,
			cmp.Or(check.Alias, check.URL),
		)
	}
}

func boolFloat(enabled bool) float64 {
	if enabled {
		return 1.0
	}
	return 0.0
}

// Describe implements Prometheus' Collector interface is used to describe metrics
func (c *ChecksCollector) Describe(ch chan<- *prometheus.Desc) {
	// log := c.Log.WithName("Describe")
	ch <- c.Enabled
}
