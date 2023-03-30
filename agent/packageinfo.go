package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type DefaultPackageInfoProvider struct{}

func (p *DefaultPackageInfoProvider) GetPackageInfo() ([]PackageInfo, error) {
	return getPackageInfo()
}

type DefaultPackageInfoReporter struct{}

func (r *DefaultPackageInfoReporter) ReportPackageInfo(config *Config, packages []PackageInfo) error {
	return reportPackageInfo(config, packages)
}

type DefaultUpgradeChecker struct{}

func (d DefaultUpgradeChecker) CheckAndPerformUpgrade(config *Config) error {
	return checkAndPerformUpgrade(config)
}

type DefaultUpgradePerformer struct{}

func (d DefaultUpgradePerformer) PerformUpgrade(autoYes bool) error {
	return performUpgrade(autoYes)
}

func getPackageInfo() ([]PackageInfo, error) {
	pm, err := getPackageManager()
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	var parseFunc func(string) []PackageInfo

	// TODO: support more package managers
	// TODO: support other operating systems
	// TODO: support multiple package managers
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

func reportPackageInfo(config *Config, packages []PackageInfo) error {
	data, err := json.Marshal(packages)
	if err != nil {
		return err
	}

	resp, err := http.Post(config.PackageEndpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func checkAndPerformUpgrade(config *Config) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.UpgradeEndpoint, nil)
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
		err = performUpgrade(true)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPackageManager() (string, error) {
	// TODO: support more package managers
	// TODO: support other operating systems
	// TODO: support multiple package managers
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
	// raw string: Inst libpulse-dev [1:15.99.1+dfsg1-1ubuntu2] (1:15.99.1+dfsg1-1ubuntu2.1 Ubuntu:22.04/jammy-updates [amd64]) []
	// format: STATUS NAME [VERSION] (NEW_VERSION CATEGORY [ARCHITECTURES]) [OTHER_INFO]

	lines := strings.Split(output, "\n")
	var packages []PackageInfo

	for _, line := range lines {
		if strings.HasPrefix(line, "Inst") {
			parts := strings.Fields(line)

			packageInfo := PackageInfo{
				Name:       parts[1],
				Version:    strings.Trim(parts[2], "[]"),
				NewVersion: strings.Trim(parts[3], "()"),
				Category:   parts[4],
				Arch:       strings.Trim(parts[5], "[]"),
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

func performUpgrade(autoYes bool) error {
	fmt.Println("Upgrading packages...")

	pm, err := getPackageManager()
	if err != nil {
		log.Printf("Error determining package manager: %v\n", err)
		return err
	}

	// TODO: support more package managers
	// TODO: support other operating systems
	// TODO: support multiple package managers
	// TODO: when in daemon mode, no user interaction should be required
	var cmdArgs []string
	switch pm {
	case "apt-get":
		cmdArgs = append(cmdArgs, pm, "upgrade")
		if autoYes {
			cmdArgs = append(cmdArgs, "-y")
		}
	case "dnf", "yum":
		cmdArgs = append(cmdArgs, pm, "upgrade")
		if autoYes {
			cmdArgs = append(cmdArgs, "-y")
		}
	default:
		err := errors.New("unsupported package manager")
		log.Printf("%v\n", err)
		return err
	}

	log.Printf("Running %v\n", strings.Join(cmdArgs, " "))

	// Run the command in a new interactive shell
	cmd := exec.Command("sudo", append([]string{"-i", "--"}, cmdArgs...)...)

	// Connect the command's stdin, stdout, and stderr to the current process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Error upgrading packages: %v\n", err)
		return err
	}

	fmt.Println("Upgrade complete.")
	return nil
}
