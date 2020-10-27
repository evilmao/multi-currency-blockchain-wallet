package monitor

import (
	"fmt"
	"sync"
	"time"

	"upex-wallet/wallet-base/newbitx/misclib/log"

	"github.com/DataDog/datadog-go/statsd"
)

var (
	defaultStatsdReporter *StatsdReporter
)

// MetricsTags is tags of metrics.
type MetricsTags map[string]string

func (tags MetricsTags) slice() []string {
	if len(tags) == 0 {
		return nil
	}

	slice := make([]string, 0, len(tags))
	for k, v := range tags {
		slice = append(slice, k+":"+v)
	}
	return slice
}

// StatsdReporter represents a statsd client that report metrics every interval.
type StatsdReporter struct {
	sync.Mutex
	c         *statsd.Client
	quit      chan struct{}
	reporters []func() error
}

// NewStatsdReporter returns a report for sending messages to dogstatsd.
func NewStatsdReporter(addr, namespace string, tags []string) *StatsdReporter {
	// Create the client
	c, err := statsd.New(addr)
	if err != nil {
		log.Fatal(err)
	}
	// Prefix every metric with the app name
	c.Namespace = fmt.Sprintf("%s.", namespace)
	// Send the EC2 availability zone as a tag with every metric
	c.Tags = append(c.Tags, tags...)

	defaultStatsdReporter = &StatsdReporter{
		c:    c,
		quit: make(chan struct{}),
	}

	return defaultStatsdReporter
}

// Start starts a statsd schedule.
func (s *StatsdReporter) Start() {
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.reportMetrics()
			case <-s.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *StatsdReporter) addMetrics(reporter func() error) {
	s.Lock()
	defer s.Unlock()

	s.reporters = append(s.reporters, reporter)
}

// Gauge measures the value of a metric at a particular time.
func (s *StatsdReporter) Gauge(name string, value float64, tags MetricsTags) {
	s.addMetrics(func() error {
		return s.c.Gauge(name, value, tags.slice(), 1)
	})
}

// Count tracks how many times something happened per second.
func (s *StatsdReporter) Count(name string, value int64, tags MetricsTags) {
	s.addMetrics(func() error {
		return s.c.Count(name, value, tags.slice(), 1)
	})
}

// Histogram tracks the statistical distribution of a set of values on each host.
func (s *StatsdReporter) Histogram(name string, value float64, tags MetricsTags) {
	s.addMetrics(func() error {
		return s.c.Histogram(name, value, tags.slice(), 1)
	})
}

// Close closes the statsd reporter.
func (s *StatsdReporter) Close() {
	s.quit <- struct{}{}
}

func (s *StatsdReporter) reportMetrics() {
	s.Lock()
	defer s.Unlock()

	for _, reporter := range s.reporters {
		err := reporter()
		_ = err
	}
	s.reporters = nil
}

// MetricsGauge measures the value of a metric at a particular time.
func MetricsGauge(name string, value float64, tags MetricsTags) {
	if defaultStatsdReporter == nil {
		return
	}
	defaultStatsdReporter.Gauge(name, value, tags)
}

// MetricsCount tracks how many times something happened per second.
func MetricsCount(name string, value int64, tags MetricsTags) {
	if defaultStatsdReporter == nil {
		return
	}
	defaultStatsdReporter.Count(name, value, tags)
}

// MetricsHistogram tracks the statistical distribution of a set of values on each host.
func MetricsHistogram(name string, value float64, tags MetricsTags) {
	if defaultStatsdReporter == nil {
		return
	}
	defaultStatsdReporter.Histogram(name, value, tags)
}
