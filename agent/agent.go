package agent

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bluet/syspkg"

	"github.com/bluet/redt-agent/utils"
)

type PackageInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	NewVersion string `json:"new_version,omitempty"`
	Category   string `json:"category,omitempty"`
	Arch       string `json:"arch,omitempty"`
}

// TelemetryData contains the collected telemetry information
type TelemetryData struct {
	CPUUsage      float64     `json:"cpu_usage"`
	MemoryUsage   float64     `json:"memory_usage"`
	OSInfo        string      `json:"os_info"`
	CurrentUser   string      `json:"current_user"`
	LoggedInUsers []string    `json:"logged_in_user"`
	DiskUsage     []DiskUsage `json:"disk_usage"`
}

// struct to hold all disk usage information
type DiskUsage struct {
	Path        string  `json:"path"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type TelemetryDataProvider interface {
	CollectTelemetryData(config *Config) (TelemetryData, error)
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

func RunDaemon() {
	log.Printf("Starting RedT agent at %s\n", utils.CurrentTimestamp())

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	log.Printf("Loaded configuration: %v", config)
	log.Printf("Checking for telemetry in %s\n", config.PollInterval)
	log.Printf("Checking for package updates in %s\n", config.UpgradeCheckPeriod)

	lastUpgradeCheck := time.Now().Add(-config.UpgradeCheckPeriod)

	ticker := time.NewTicker(config.PollInterval)
	for range ticker.C {
		handleTelemetry(config, &DefaultTelemetryDataProvider{}, &DefaultTelemetryDataSender{})
		lastUpgradeCheck = handlePackageInfo(config, &DefaultPackageInfoProvider{}, &DefaultPackageInfoReporter{}, lastUpgradeCheck, &DefaultUpgradeChecker{})
	}
}

func RunShowMetrics() error {
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %v", err)
	}
	log.Printf("Loaded configuration: %v", config)

	// Print system info
	// Create an instance of DefaultTelemetryDataProvider
	telemetryDataProvider := DefaultTelemetryDataProvider{}

	// Call CollectTelemetryData on the instance
	data, err := telemetryDataProvider.CollectTelemetryData(config)
	if err != nil {
		return fmt.Errorf("Error collecting telemetry data: %v", err)
	}
	fmt.Println("System Information:")
	fmt.Printf("- OS Info: %s\n", data.OSInfo)
	fmt.Printf("- CPU Usage: %.2f%%\n", data.CPUUsage)
	fmt.Printf("- Memory Usage: %.2f%%\n", data.MemoryUsage)
	fmt.Printf("- Current User: %s\n", data.CurrentUser)
	fmt.Printf("- Logged In User: %s\n", data.LoggedInUsers)
	fmt.Printf("- Disk Usage:\n")
	for _, diskUsage := range data.DiskUsage {
		fmt.Printf("  - %s: %.2f%%\n", diskUsage.Path, diskUsage.UsedPercent)
	}

	// Print upgradable packages
	fmt.Println("Checking for upgradable packages...")
	// Create instances of DefaultUpgradeChecker and DefaultUpgradePerformer
	// packageInfo := DefaultPackageInfoProvider{}
	// upgradablePackages, err := packageInfo.GetPackageInfo()
	pms, err := syspkg.NewPackageManager()
	if err != nil {
		fmt.Printf("Error while initializing package managers: %v", err)
		os.Exit(1)
	}

	var upgradablePackages []syspkg.PackageInfo
	for _, pm := range pms {
		pkgs, err := pm.ListUpgradable()
		if err != nil {
			return fmt.Errorf("Error checking for upgradable packages: %T: %v", pm, err)
		}
		upgradablePackages = append(upgradablePackages, pkgs...)
	}

	if len(upgradablePackages) > 0 {
		fmt.Println("Upgradable packages:")
		for _, pkg := range upgradablePackages {
			fmt.Printf("%s: %s %s -> %s (%s)\n", pkg.PackageManager, pkg.Name, pkg.Version, pkg.NewVersion, pkg.Status)
		}
	} else {
		fmt.Println("No upgradable packages found.")
	}

	return nil
}

func RunSysup(autoYes bool) error {
	// Call CheckUpgradablePackages on the upgradeChecker instance
	// config, err := LoadConfig()
	// if err != nil {
	// 	return fmt.Errorf("Error loading configuration: %v", err)
	// }

	pms, err := syspkg.NewPackageManager()
	if err != nil {
		fmt.Printf("Error while initializing package managers: %v", err)
		// fmt.Println("Error:", err)
		os.Exit(1)
	}

	err = RunShowMetrics()
	if err != nil {
		return err
	}

	// Create instances of DefaultUpgradeChecker and DefaultUpgradePerformer
	// upgradePerformer := DefaultUpgradePerformer{}

	if !autoYes {
		fmt.Print("Do you want to perform the upgrade? (Y/n) ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "" {
			return nil
		}
	}

	// Call PerformUpgrade on the upgradePerformer instance
	// err = upgradePerformer.PerformUpgrade(autoYes)
	fmt.Println("Performing package upgrade...")

	for _, pm := range pms {
		err := pm.Upgrade()
		if err != nil {
			fmt.Printf("Error performing system upgrade: %v", err)
		}
	}

	fmt.Println("System upgrade completed successfully.")
	return nil
}

func handleTelemetry(config *Config, telemetryDataProvider TelemetryDataProvider, telemetryDataSender TelemetryDataSender) {

	telemetryData, err := telemetryDataProvider.CollectTelemetryData(config)
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
		log.Printf("Package info: %v", packages)

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
