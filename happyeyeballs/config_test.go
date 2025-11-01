package happyeyeballs

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedConfig Config
	}{
		{
			name:    "default_config",
			envVars: map[string]string{},
			expectedConfig: Config{
				Enabled:         false,
				ResolutionDelay: DefaultResolutionDelay,
				ConnectionDelay: DefaultConnectionDelay,
				MetricsEnabled:  false,
				VerboseLogging:  false,
			},
		},
		{
			name: "enabled_config",
			envVars: map[string]string{
				"LETSDANE_HAPPY_EYEBALLS": "true",
			},
			expectedConfig: Config{
				Enabled:         true,
				ResolutionDelay: DefaultResolutionDelay,
				ConnectionDelay: DefaultConnectionDelay,
				MetricsEnabled:  true,
				VerboseLogging:  false,
			},
		},
		{
			name: "custom_delays",
			envVars: map[string]string{
				"LETSDANE_HAPPY_EYEBALLS":      "true",
				"LETSDANE_HE_RESOLUTION_DELAY": "100",
				"LETSDANE_HE_CONNECTION_DELAY": "300",
			},
			expectedConfig: Config{
				Enabled:         true,
				ResolutionDelay: 100 * time.Millisecond,
				ConnectionDelay: 300 * time.Millisecond,
				MetricsEnabled:  true,
				VerboseLogging:  false,
			},
		},
		{
			name: "delay_clamping",
			envVars: map[string]string{
				"LETSDANE_HAPPY_EYEBALLS":      "true",
				"LETSDANE_HE_CONNECTION_DELAY": "50",
			},
			expectedConfig: Config{
				Enabled:         true,
				ResolutionDelay: DefaultResolutionDelay,
				ConnectionDelay: MinConnectionDelay,
				MetricsEnabled:  true,
				VerboseLogging:  false,
			},
		},
		{
			name: "all_features_enabled",
			envVars: map[string]string{
				"LETSDANE_HAPPY_EYEBALLS": "true",
				"LETSDANE_HE_METRICS":     "true",
				"LETSDANE_HE_VERBOSE":     "true",
			},
			expectedConfig: Config{
				Enabled:         true,
				ResolutionDelay: DefaultResolutionDelay,
				ConnectionDelay: DefaultConnectionDelay,
				MetricsEnabled:  true,
				VerboseLogging:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			cfg := LoadConfigFromEnv()

			if cfg.Enabled != tt.expectedConfig.Enabled {
				t.Errorf("Enabled: got %v, want %v", cfg.Enabled, tt.expectedConfig.Enabled)
			}
			if cfg.ResolutionDelay != tt.expectedConfig.ResolutionDelay {
				t.Errorf("ResolutionDelay: got %v, want %v", cfg.ResolutionDelay, tt.expectedConfig.ResolutionDelay)
			}
			if cfg.ConnectionDelay != tt.expectedConfig.ConnectionDelay {
				t.Errorf("ConnectionDelay: got %v, want %v", cfg.ConnectionDelay, tt.expectedConfig.ConnectionDelay)
			}
			if cfg.MetricsEnabled != tt.expectedConfig.MetricsEnabled {
				t.Errorf("MetricsEnabled: got %v, want %v", cfg.MetricsEnabled, tt.expectedConfig.MetricsEnabled)
			}
			if cfg.VerboseLogging != tt.expectedConfig.VerboseLogging {
				t.Errorf("VerboseLogging: got %v, want %v", cfg.VerboseLogging, tt.expectedConfig.VerboseLogging)
			}
		})
	}
}
