package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

type DefaultPackageInfoProvider struct{}

func (p *DefaultPackageInfoProvider) GetPackageInfo() ([]PackageInfo, error) {
	return getPackageInfo()
}

type DefaultPackageInfoReporter struct{}

func (r *DefaultPackageInfoReporter) ReportPackageInfo(packages []PackageInfo) error {
	return reportPackageInfo(packages)
}

func getPackageInfo() ([]PackageInfo, error) {
	pm, err := getPackageManager()
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	var parseFunc func(string) []PackageInfo

	switch pm {
	case "apt-get":
		cmd = exec.Command(pm, "upgrade", "-s")
		parseFunc = parseAptGetOutput
	case "dnf", "yum":
		cmd = exec.Command(pm, "check-update")
		parseFunc = parseDnfYumOutput
	default:
		return nil, fmt.Errorf("unsupported package manager")
	}

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseFunc(string(out)), nil
}

func reportPackageInfo(packages []PackageInfo) error {
	data, err := json.Marshal(packages)
	if err != nil {
		return err
	}

	resp, err := http.Post(packageEndpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func checkAndPerformUpgrade() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", upgradeEndpoint, nil)
	if err != nil {
		log.Printf("Error creating upgrade request: %v\n", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error checking for upgrades: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		err = performUpgrade()
		if err != nil {
			return err
		}
	}

	return nil
}

func getPackageManager() (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("unsupported operating system")
	}

	_, err := exec.LookPath("apt-get")
	if err == nil {
		return "apt-get", nil
	}

	_, err = exec.LookPath("dnf")
	if err == nil {
		return "dnf", nil
	}

	_, err = exec.LookPath("yum")
	if err == nil {
		return "yum", nil
	}

	return "", fmt.Errorf("package manager not found")
}

func parseAptGetOutput(output string) []PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []PackageInfo

	for _, line := range lines {
		if strings.HasPrefix(line, "Inst") {
			parts := strings.Fields(line)

			packageInfo := PackageInfo{
				Name:       parts[1],
				Version:    parts[2],
				NewVersion: parts[4],
			}

			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func parseDnfYumOutput(output string) []PackageInfo {
	lines := strings.Split(output, "\n")
	var packages []PackageInfo

	for _, line := range lines {
		parts := strings.Fields(line)

		if len(parts) >= 3 {
			packageInfo := PackageInfo{
				Name:       parts[0],
				Version:    parts[1],
				NewVersion: parts[2],
			}

			packages = append(packages, packageInfo)
		}
	}

	return packages
}

func performUpgrade() error {
	fmt.Println("Upgrading packages...")

	pm, err := getPackageManager()
	if err != nil {
		log.Printf("Error determining package manager: %v\n", err)
		return err
	}

	var cmd *exec.Cmd

	switch pm {
	case "apt-get":
		cmd = exec.Command("sudo", pm, "upgrade", "-y")
	case "dnf", "yum":
		cmd = exec.Command("sudo", pm, "upgrade", "-y")
	default:
		err := errors.New("unsupported package manager")
		log.Printf("%v\n", err)
		return err
	}

	err = cmd.Run()
	if err != nil {
		log.Printf("Error upgrading packages: %v\n", err)
		return err
	}

	fmt.Println("Upgrade complete.")
	return nil
}
