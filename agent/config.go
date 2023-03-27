package agent

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	BackendURL         string
	TelemetryEndpoint  string
	PackageEndpoint    string
	UpgradeEndpoint    string
	PollInterval       time.Duration
	UpgradeCheckPeriod time.Duration
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	backendURL := viper.GetString("backendURL")
	pollInterval := viper.GetDuration("pollInterval") * time.Second
	upgradeCheckPeriod := viper.GetDuration("upgradeCheckPeriod") * time.Minute

	return &Config{
		BackendURL:         backendURL,
		TelemetryEndpoint:  backendURL + "/telemetry",
		PackageEndpoint:    backendURL + "/packages",
		UpgradeEndpoint:    backendURL + "/upgrade",
		PollInterval:       pollInterval,
		UpgradeCheckPeriod: upgradeCheckPeriod,
	}, nil
}
