package agent

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	BackendURL         string `yaml:"backend_url"`
	TelemetryEndpoint  string
	PackageEndpoint    string
	UpgradeEndpoint    string
	PollInterval       time.Duration `yaml:"poll_interval"`
	UpgradeCheckPeriod time.Duration `yaml:"upgrade_check_period"`
	Token              string        `yaml:"token"`
	Hostname           string        `yaml:"hostname"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	backendURL := viper.GetString("backend_url")
	pollInterval := viper.GetDuration("poll_interval") * time.Second
	upgradeCheckPeriod := viper.GetDuration("upgrade_check_period") * time.Minute
	token := viper.GetString("token")
	hostname := viper.GetString("hostname")

	return &Config{
		BackendURL:         backendURL,
		TelemetryEndpoint:  backendURL + "/telemetry",
		PackageEndpoint:    backendURL + "/packages",
		UpgradeEndpoint:    backendURL + "/upgrade",
		PollInterval:       pollInterval,
		UpgradeCheckPeriod: upgradeCheckPeriod,
		Token:              token,
		Hostname:           hostname,
	}, nil
}
