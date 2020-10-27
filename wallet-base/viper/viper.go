package viper

import (
	"time"

	"github.com/spf13/viper"
)

func Get(key string) interface{} {
	return viper.Get(key)
}

func GetInt64(key string, defaultValue int64) int64 {
	if viper.IsSet(key) {
		return viper.GetInt64(key)
	}
	return defaultValue
}

func GetFloat64(key string, defaultValue float64) float64 {
	if viper.IsSet(key) {
		return viper.GetFloat64(key)
	}
	return defaultValue
}

func GetString(key string, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}

func GetStringSlice(key string, defaultValue []string) []string {
	if viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}
	return defaultValue
}

func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if viper.IsSet(key) {
		v := viper.GetString(key)
		if duration, err := time.ParseDuration(v); err == nil {
			return duration
		}
		return 0
	}
	return defaultValue
}

func GetBool(key string, defaultValue bool) bool {
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}
	return defaultValue
}

func GetStringMapString(key string, defaultValue map[string]string) map[string]string {
	if viper.IsSet(key) {
		return viper.GetStringMapString(key)
	}
	return defaultValue
}

func GetStringMap(key string, defaultValue map[string]interface{}) map[string]interface{} {
	if viper.IsSet(key) {
		return viper.GetStringMap(key)
	}
	return defaultValue
}
