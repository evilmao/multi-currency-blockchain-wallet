package utils

import (
	"time"

	"github.com/spf13/viper"
)

func GetInt64(key string, defaultValue int64) int64 {
	var (
		value int64
	)
	if value = viper.GetInt64(key); value == 0 {
		return defaultValue
	}
	return value
}

func GetFloat64(key string, defaultValue float64) float64 {
	var (
		value float64
	)
	if value = viper.GetFloat64(key); value == 0 {
		return defaultValue
	}
	return value
}

func GetString(key string, defaultValue string) string {
	var (
		value string
	)
	if value = viper.GetString(key); value == "" {
		return defaultValue
	}
	return value
}

func GetStringSlice(key string, defaultValue []string) []string {
	var (
		value []string
	)
	if value = viper.GetStringSlice(key); len(value) == 0 {
		return defaultValue
	}
	return value
}

func GetDuration(key string, defaultValue time.Duration) time.Duration {
	var (
		value string
	)
	if value = viper.GetString(key); value == "" {
		return defaultValue
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	return defaultValue
}

func GetBool(key string, defaultValue bool) bool {
	var (
		value bool
	)
	if value = viper.GetBool(key); value == false {
		return defaultValue
	}
	return value
}

func GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	var (
		value = make(map[string]string)
	)
	if value = viper.GetStringMapString(key); value == nil {
		return defaultValue
	}
	return value
}

func GetStringMap(key string, defaultValue map[string]interface{}) map[string]interface{} {
	var (
		value = make(map[string]interface{})
	)
	if value = viper.GetStringMap(key); value == nil {
		return defaultValue
	}
	return value
}
