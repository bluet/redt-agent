package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/user"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type DefaultTelemetryDataProvider struct{}

func (d DefaultTelemetryDataProvider) CollectTelemetryData() (TelemetryData, error) {
	return collectTelemetryData()
}

type DefaultTelemetryDataSender struct{}

func (d DefaultTelemetryDataSender) SendTelemetryData(config *Config, data TelemetryData) error {
	return sendTelemetryData(config, data)
}

func collectTelemetryData() (TelemetryData, error) {
	var data TelemetryData

	// Collect CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return data, fmt.Errorf("failed to collect CPU usage: %v", err)
	}
	data.CPUUsage = cpuPercent[0]

	// Collect memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return data, fmt.Errorf("failed to collect memory usage: %v", err)
	}
	data.MemoryUsage = memInfo.UsedPercent

	// Collect disk usage
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return data, fmt.Errorf("failed to collect disk usage: %v", err)
	}
	data.DiskUsage = diskInfo.UsedPercent

	// Collect OS info
	hostInfo, err := host.Info()
	if err != nil {
		return data, fmt.Errorf("failed to collect OS info: %v", err)
	}
	data.OSInfo = fmt.Sprintf("%s %s %s", hostInfo.OS, hostInfo.Platform, hostInfo.PlatformVersion)

	// Collect logged-in user
	currentUser, err := user.Current()
	if err != nil {
		return data, fmt.Errorf("failed to collect logged-in user: %v", err)
	}
	data.LoggedInUser = currentUser.Username

	return data, nil
}

func sendTelemetryData(config *Config, data TelemetryData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry data: %v", err)
	}

	resp, err := http.Post(config.TelemetryEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send telemetry data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("backend responded with status %d", resp.StatusCode)
	}

	return nil
}
