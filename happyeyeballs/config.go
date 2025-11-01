package happyeyeballs

import (
	"os"
	"strconv"
	"time"
)

const (
	DefaultResolutionDelay     = 50 * time.Millisecond
	DefaultConnectionDelay     = 250 * time.Millisecond
	MinConnectionDelay         = 100 * time.Millisecond
	MaxConnectionDelay         = 2 * time.Second
)

type Config struct {
	Enabled         bool
	ResolutionDelay time.Duration
	ConnectionDelay time.Duration
	MetricsEnabled  bool
	VerboseLogging  bool
}

func LoadConfigFromEnv() *Config {
	cfg := &Config{
		Enabled:         parseBoolEnv("LETSDANE_HAPPY_EYEBALLS", false),
		ResolutionDelay: parseDurationEnv("LETSDANE_HE_RESOLUTION_DELAY", DefaultResolutionDelay),
		ConnectionDelay: parseDurationEnv("LETSDANE_HE_CONNECTION_DELAY", DefaultConnectionDelay),
		VerboseLogging:  parseBoolEnv("LETSDANE_HE_VERBOSE", false),
	}

	cfg.MetricsEnabled = parseBoolEnv("LETSDANE_HE_METRICS", cfg.Enabled)

	if cfg.ConnectionDelay < MinConnectionDelay {
		cfg.ConnectionDelay = MinConnectionDelay
	}
	if cfg.ConnectionDelay > MaxConnectionDelay {
		cfg.ConnectionDelay = MaxConnectionDelay
	}

	return cfg
}

func parseBoolEnv(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return b
}

func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	ms, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultValue
	}

	return time.Duration(ms) * time.Millisecond
}
