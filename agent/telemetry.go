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
	"golang.org/x/exp/slices"
)

type DefaultTelemetryDataProvider struct{}

func (d DefaultTelemetryDataProvider) CollectTelemetryData(config *Config) (TelemetryData, error) {
	return collectTelemetryData(config)
}

type DefaultTelemetryDataSender struct{}

func (d DefaultTelemetryDataSender) SendTelemetryData(config *Config, data TelemetryData) error {
	return sendTelemetryData(config, data)
}

func collectTelemetryData(config *Config) (TelemetryData, error) {
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

	// Collect disk usage of all mounted disks
	partitions, err := disk.Partitions(false)
	if err != nil {
		return data, fmt.Errorf("failed to collect disk usage: %v", err)
	}
	for _, partition := range partitions {
		// only collect disk usage for:
		// fstype included in the config.DiskUsage.FSTypes or mountpoint included in config.DiskUsage.MountPoints
		if !slices.Contains(config.DiskUsage.FSTypes, partition.Fstype) || !slices.Contains(config.DiskUsage.Mountpoints, partition.Mountpoint) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return data, fmt.Errorf("failed to collect disk usage: %v", err)
		}
		data.DiskUsage = append(data.DiskUsage, DiskUsage{
			Path:        partition.Mountpoint,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	// Collect OS info
	hostInfo, err := host.Info()
	if err != nil {
		return data, fmt.Errorf("failed to collect OS info: %v", err)
	}
	data.OSInfo = fmt.Sprintf("%s %s %s", hostInfo.OS, hostInfo.Platform, hostInfo.PlatformVersion)

	// Collect the user which the agent is running as
	currentUser, err := user.Current()
	if err != nil {
		return data, fmt.Errorf("failed to collect logged-in user: %v", err)
	}
	data.CurrentUser = currentUser.Username

	// Collect all logged-in users
	onlineUsers, err := host.Users()
	if err != nil {
		return data, fmt.Errorf("failed to collect logged-in users: %v", err)
	}
	for _, user := range onlineUsers {
		data.LoggedInUsers = append(data.LoggedInUsers, user.User)
	}

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
