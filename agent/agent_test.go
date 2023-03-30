package agent

import (
	"errors"
	"testing"
	"time"
)

type MockTelemetryDataProvider struct {
	data TelemetryData
	err  error
}

func (m MockTelemetryDataProvider) CollectTelemetryData(config *Config) (TelemetryData, error) {
	return m.data, m.err
}

type MockTelemetryDataSender struct {
	err error
}

func (m *MockTelemetryDataSender) SendTelemetryData(config *Config, data TelemetryData) error {
	return m.err
}

type MockPackageInfoProvider struct {
	packages []PackageInfo
	err      error
}

func (m MockPackageInfoProvider) GetPackageInfo() ([]PackageInfo, error) {
	return m.packages, m.err
}

type MockPackageInfoReporter struct {
	err error
}

func (m *MockPackageInfoReporter) ReportPackageInfo(config *Config, packages []PackageInfo) error {
	return m.err
}

type MockUpgradeChecker struct {
	err error
}

func (m *MockUpgradeChecker) CheckAndPerformUpgrade(config *Config) error {
	return m.err
}

func getTestConfig() *Config {
	return &Config{
		BackendURL:         "https://example.com/api",
		TelemetryEndpoint:  "https://example.com/api/telemetry",
		PackageEndpoint:    "https://example.com/api/packages",
		UpgradeEndpoint:    "https://example.com/api/upgrade",
		PollInterval:       60 * time.Second,
		UpgradeCheckPeriod: 5 * time.Minute,
	}
}

func TestHandleTelemetry(t *testing.T) {
	testConfig := getTestConfig()
	tests := []struct {
		name                  string
		telemetryDataProvider MockTelemetryDataProvider
		telemetryDataSender   *MockTelemetryDataSender
	}{
		{
			name: "Successful telemetry data collection and send",
			telemetryDataProvider: MockTelemetryDataProvider{
				data: TelemetryData{},
				err:  nil,
			},
			telemetryDataSender: &MockTelemetryDataSender{
				err: nil,
			},
		},
		{
			name: "Error in telemetry data collection",
			telemetryDataProvider: MockTelemetryDataProvider{
				data: TelemetryData{},
				err:  errors.New("error collecting telemetry data"),
			},
			telemetryDataSender: &MockTelemetryDataSender{
				err: nil,
			},
		},
		{
			name: "Error in telemetry data send",
			telemetryDataProvider: MockTelemetryDataProvider{
				data: TelemetryData{},
				err:  nil,
			},
			telemetryDataSender: &MockTelemetryDataSender{
				err: errors.New("error sending telemetry data"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleTelemetry(testConfig, tt.telemetryDataProvider, tt.telemetryDataSender)
		})
	}
}

func TestHandlePackageInfo(t *testing.T) {
	testConfig := getTestConfig()
	tests := []struct {
		name                string
		packageInfoProvider MockPackageInfoProvider
		packageInfoReporter *MockPackageInfoReporter
		upgradeChecker      *MockUpgradeChecker
		shouldUpdate        bool
		wantError           bool
	}{
		{
			name: "Successful package info retrieval and reporting",
			packageInfoProvider: MockPackageInfoProvider{
				packages: []PackageInfo{},
				err:      nil,
			},
			packageInfoReporter: &MockPackageInfoReporter{
				err: nil,
			},
			upgradeChecker: &MockUpgradeChecker{
				err: nil,
			},
			shouldUpdate: true,
			wantError:    false,
		},
		{
			name: "Error in package info retrieval",
			packageInfoProvider: MockPackageInfoProvider{
				packages: []PackageInfo{},
				err:      errors.New("error retrieving package info"),
			},
			packageInfoReporter: &MockPackageInfoReporter{
				err: nil,
			},
			upgradeChecker: &MockUpgradeChecker{
				err: nil,
			},
			shouldUpdate: false,
			wantError:    true,
		},
		{
			name: "Error in package info reporting",
			packageInfoProvider: MockPackageInfoProvider{
				packages: []PackageInfo{},
				err:      nil,
			},
			packageInfoReporter: &MockPackageInfoReporter{
				err: errors.New("error reporting package info"),
			},
			upgradeChecker: &MockUpgradeChecker{
				err: nil,
			},
			shouldUpdate: false,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastUpgradeCheck := time.Now().Add(-testConfig.UpgradeCheckPeriod)
			newUpgradeCheck := handlePackageInfo(testConfig, tt.packageInfoProvider, tt.packageInfoReporter, lastUpgradeCheck, tt.upgradeChecker)
			if tt.shouldUpdate && time.Since(newUpgradeCheck) >= testConfig.UpgradeCheckPeriod {
				t.Errorf("Expected lastUpgradeCheck to be updated after upgradeCheckPeriod, but it was not.")
			}

			if !tt.shouldUpdate && time.Since(newUpgradeCheck) < testConfig.UpgradeCheckPeriod {
				t.Errorf("Expected lastUpgradeCheck not to be updated, but it was.")
			}
		})
	}
}
