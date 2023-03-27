package agent

import (
	"log"
	"time"

	"github.com/bluet/redt-agent/utils"
)

type PackageInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	NewVersion string `json:"new_version,omitempty"`
}

// TelemetryData contains the collected telemetry information
type TelemetryData struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	OSInfo       string  `json:"os_info"`
	LoggedInUser string  `json:"logged_in_user"`
}

type TelemetryDataProvider interface {
	CollectTelemetryData() (TelemetryData, error)
}

type TelemetryDataSender interface {
	SendTelemetryData(config *Config, data TelemetryData) error
}

type PackageInfoProvider interface {
	GetPackageInfo() ([]PackageInfo, error)
}

type PackageInfoReporter interface {
	ReportPackageInfo(config *Config, packages []PackageInfo) error
}

type UpgradeChecker interface {
	CheckAndPerformUpgrade(config *Config) error
}

type UpgradePerformer interface {
	PerformUpgrade() error
}

func Run() {
	log.Printf("Starting RedT agent at %s\n", utils.CurrentTimestamp())

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	lastUpgradeCheck := time.Now().Add(-config.UpgradeCheckPeriod)

	ticker := time.NewTicker(config.PollInterval)
	for range ticker.C {
		handleTelemetry(config, &DefaultTelemetryDataProvider{}, &DefaultTelemetryDataSender{})
		lastUpgradeCheck = handlePackageInfo(config, &DefaultPackageInfoProvider{}, &DefaultPackageInfoReporter{}, lastUpgradeCheck, &DefaultUpgradeChecker{})
	}
}

func handleTelemetry(config *Config, telemetryDataProvider TelemetryDataProvider, telemetryDataSender TelemetryDataSender) {

	telemetryData, err := telemetryDataProvider.CollectTelemetryData()
	if err != nil {
		log.Printf("Error collecting telemetry data: %v", err)
	} else {
		err = telemetryDataSender.SendTelemetryData(config, telemetryData)
		if err != nil {
			log.Printf("Error sending telemetry data: %v", err)
		}
	}
}

func handlePackageInfo(config *Config, provider PackageInfoProvider, reporter PackageInfoReporter, lastUpgradeCheck time.Time, upgradeChecker UpgradeChecker) time.Time {
	if time.Since(lastUpgradeCheck) >= config.UpgradeCheckPeriod {
		packages, err := provider.GetPackageInfo()
		if err != nil {
			log.Printf("Error getting package info: %v", err)
		} else {
			err = reporter.ReportPackageInfo(config, packages)
			if err != nil {
				log.Printf("Error reporting package info: %v", err)
			} else {
				err = upgradeChecker.CheckAndPerformUpgrade(config)
				if err != nil {
					log.Printf("Error checking and performing upgrade: %v", err)
				}
				lastUpgradeCheck = time.Now()
			}
		}
	}
	return lastUpgradeCheck
}
