package collector

import (
	"cmp"
	"sync"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"

	"kastelo.dev/updown-exporter/updown"
)

// MetricsCollector is a type that represents updown Check Metrics
type MetricsCollector struct {
	System       System
	Client       *updown.Client
	Log          logr.Logger
	ResponseTime *prometheus.Desc
}

// NewMetricsCollector is a function that returns a new MetricsCollector
func NewMetricsCollector(s System, client *updown.Client, log logr.Logger) *MetricsCollector {
	subsystem := "metrics"
	return &MetricsCollector{
		System: s,
		Client: client,
		Log:    log,
		ResponseTime: prometheus.NewDesc(
			prometheus.BuildFQName(s.Namespace, subsystem, "response_time_seconds"),
			"check metrics response times (seconds)",
			[]string{
				"url",
				"alias",
			},
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.Log.WithName("Collect")

	// In order to get Metrics, need a Check
	checks, err := c.Client.GetChecks()
	if err != nil {
		log.Info("unable to get Checks")
		return
	}

	var wg sync.WaitGroup
	for _, check := range checks {
		log := log.WithValues("URL", check.URL)

		if check.Token == "" {
			log.Info("unable to obtain token for Check")
			return
		}

		metrics, err := c.Client.GetCheckMetrics(check.Token)
		if err != nil {
			log.Error(err, "unable to read metrics for Check")
			return
		}

		respTime := metrics.Requests.ByResponseTime
		ch <- prometheus.MustNewConstHistogram(
			c.ResponseTime,
			// updown doesn't provide values for above 4s (i.e. Infinity)
			// website only permits maxium value of 2s so I assume 4s is intended to represent "all else"
			// Assuming that Under4000 is effectively infinity and using it as the value for count
			uint64(respTime.Under4000),
			// updown doesn't provide a value for the sum of values
			0.0,
			// Convert the struct into a map of buckets
			respTime.ToBuckets(),
			check.URL,
			cmp.Or(check.Alias, check.URL),
		)
	}
	wg.Wait()
}

// Describe implements Prometheus' Collector interface and is used to describe metrics
func (c *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ResponseTime
}
