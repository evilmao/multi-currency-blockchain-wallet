package util

import (
	"fmt"
	"time"

	"upex-wallet/wallet-base/newbitx/misclib/log"
)

// DeferTimeCost calculates time cost.
func DeferTimeCost(h func(time.Duration)) func() {
	start := time.Now()
	return func() {
		h(time.Now().Sub(start))
	}
}

// TimeCost calculates time cost.
func TimeCost(f func()) (cost time.Duration) {
	defer DeferTimeCost(func(d time.Duration) {
		cost = d
	})()

	f()
	return
}

// DeferLogTimeCost logs time cost.
func DeferLogTimeCost(tag string) func() {
	return DeferTimeCost(func(d time.Duration) {
		log.Infof("%s time cost: %s", tag, d)
	})
}

// WithLogTimeCost logs time cost.
func WithLogTimeCost(tag string, f func()) {
	defer DeferLogTimeCost(tag)()
	f()
}

var (
	// ErrTimeout returns when timeout.
	ErrTimeout = fmt.Errorf("timeout")
)

// WithTimeout runs f with a timeout.
func WithTimeout(duration time.Duration, f func() error) error {
	var (
		c   = make(chan struct{})
		err error
	)

	Go("WithTimeout", func() {
		err = f()
		close(c)
	}, func(panicErr error) {
		err = panicErr
		close(c)
	})

	select {
	case <-c:
		return err
	case <-time.After(duration):
		return ErrTimeout
	}
}

// UpdateMonitor monitors for update interval.
type UpdateMonitor struct {
	originMaxInterval time.Duration
	handleTimeout     func(int64, time.Duration)
	adjustInterval    bool

	maxInterval    time.Duration
	value          int64
	lastUpdateTime time.Time
}

// NewUpdateMonitor returns a new UpdateMonitor.
func NewUpdateMonitor(maxInterval time.Duration, handleTimeout func(int64, time.Duration)) *UpdateMonitor {
	return newUpdateMonitor(maxInterval, handleTimeout, false)
}

// NewAutoAdjustUpdateMonitor returns a new auto adjust interval UpdateMonitor.
func NewAutoAdjustUpdateMonitor(maxInterval time.Duration, handleTimeout func(int64, time.Duration)) *UpdateMonitor {
	return newUpdateMonitor(maxInterval, handleTimeout, true)
}

func newUpdateMonitor(maxInterval time.Duration, handleTimeout func(int64, time.Duration), autoAdjustInterval bool) *UpdateMonitor {
	return &UpdateMonitor{
		originMaxInterval: maxInterval,
		handleTimeout:     handleTimeout,
		adjustInterval:    autoAdjustInterval,

		maxInterval:    maxInterval,
		value:          0,
		lastUpdateTime: time.Now(),
	}
}

// Update updates monitor state.
func (m *UpdateMonitor) Update(value int64) {
	if value != m.value {
		m.value = value
		m.lastUpdateTime = time.Now()
		m.maxInterval = m.originMaxInterval
	} else {
		m.Check()
	}
}

// Check checks timeout.
func (m *UpdateMonitor) Check() {
	if m.handleTimeout == nil {
		return
	}

	interval := time.Now().Sub(m.lastUpdateTime)
	if interval > m.maxInterval {
		m.handleTimeout(m.value, interval)

		if m.adjustInterval {
			m.maxInterval += m.originMaxInterval
		}
	}
}

// Value returns the value of UpdateMoniter.
func (m *UpdateMonitor) Value() int64 {
	return m.value
}
