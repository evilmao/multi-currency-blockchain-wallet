package util

import (
	"time"
)

// Try tries fn maxTimes.
func Try(maxTimes int, fn func(int) error) (err error) {
	return TryWithInterval(maxTimes, 0, fn)
}

// TryWithInterval tries fn maxTimes with interval.
func TryWithInterval(maxTimes int, interval time.Duration, fn func(int) error) (err error) {
	for i := 0; i < maxTimes; i++ {
		if err = fn(i); err == nil {
			return
		}

		time.Sleep(interval)
	}
	return
}

// TryCatch tries fn maxTimes, call catch if failed.
func TryCatch(maxTimes int, fn func(int) error, catch func(error)) (err error) {
	err = TryWithInterval(maxTimes, 0, fn)
	if err != nil {
		catch(err)
		return
	}
	return
}
