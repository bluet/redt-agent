package agent

import (
	"log"
	"time"

	"github.com/bluet/redt-agent/utils"
)

const (
	backendURL         = "https://redt.top/api"
	telemetryEndpoint  = backendURL + "/telemetry"
	packageEndpoint    = backendURL + "/packages"
	upgradeEndpoint    = backendURL + "/upgrade"
	pollInterval       = 60 * time.Second
	upgradeCheckPeriod = 5 * time.Minute
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

func Run() {
	log.Printf("Starting RedT agent at %s\n", utils.CurrentTimestamp())

	lastUpgradeCheck := time.Now().Add(-upgradeCheckPeriod)

	ticker := time.NewTicker(pollInterval)
	for range ticker.C {
		handleTelemetry()
		lastUpgradeCheck = handlePackageInfo(lastUpgradeCheck)
		checkAndPerformUpgrade()
	}
}

func handleTelemetry() {
	telemetryData, err := collectTelemetryData()
	if err != nil {
		log.Printf("Error collecting telemetry data: %v", err)
	} else {
		err = sendTelemetryData(telemetryData)
		if err != nil {
			log.Printf("Error sending telemetry data: %v", err)
		}
	}
}

func handlePackageInfo(lastUpgradeCheck time.Time) time.Time {
	if time.Since(lastUpgradeCheck) >= upgradeCheckPeriod {
		packages, err := getPackageInfo()
		if err != nil {
			log.Printf("Error getting package info: %v", err)
		} else {
			err = reportPackageInfo(packages)
			if err != nil {
				log.Printf("Error reporting package info: %v", err)
			}
			lastUpgradeCheck = time.Now()
		}
	}
	return lastUpgradeCheck
}
