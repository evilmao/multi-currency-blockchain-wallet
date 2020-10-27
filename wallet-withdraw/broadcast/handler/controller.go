package handler

import (
	"time"

	bviper "upex-wallet/wallet-base/viper"
)

const (
	_minVerifyInterval     = 1
	_defaultVerifyInterval = 3
	_maxVerifyInterval     = 60
)

type Controller struct {
	configPrefix string
}

func (c *Controller) SetConfigPrefix(prefix string) {
	c.configPrefix = prefix
}

func (c *Controller) VerifyInterval() time.Duration {
	interval := bviper.GetInt64(c.configPrefix+".verifyInterval", _defaultVerifyInterval)
	interval = normalizeInt64(interval, _minVerifyInterval, _maxVerifyInterval)
	return time.Duration(interval) * time.Second
}

func normalizeInt64(v, min, max int64) int64 {
	if min > max {
		min, max = max, min
	}

	if v < min {
		return min
	}

	if v > max {
		return max
	}

	return v
}
