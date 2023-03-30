# RedT Agent

[![Status](https://img.shields.io/badge/status-active-success.svg)](https://github.com/bluet/redt-agent/)
[![GitHub Issues](https://img.shields.io/github/issues/bluet/redt-agent.svg)](https://github.com/bluet/redt-agent/issues)
[![GitHub Pull Requests](https://img.shields.io/github/issues-pr/bluet/redt-agent.svg)](https://github.com/bluet/redt-agent/pulls)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](/LICENSE)

`redt-agent` is a lightweight, extensible agent that collects telemetry data, and package information, and automatically checks for software upgrades for your application. It communicates with a backend service to report collected data and receive upgrade instructions.

## Features

- Collects telemetry data
- Reports package information
- Automatically checks for software upgrades
- Configurable polling intervals and upgrade check periods

## Getting Started

These instructions will help you set up and configure the `redt-agent` for your application.

### Prerequisites

- Go 1.16 or later (1.20 preferred)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/bluet/redt-agent.git

```

1. Change to the project directory:

```bash
cd redt-agent
```

1. Build the project:

```bash
make
```

#### Run as a standalone application or command line tool

```bash
# show info (cpu, memory, disk, network, process, package)
./bin/redt-agent
# show info then do system package upgrade (with prompt before upgrade)
./bin/redt-agent sysup
# show info then do system package upgrade (without prompt before upgrade)
./bin/redt-agent sysup -y
```

#### Run as a daemon

```bash
./bin/redt-agent -d
```

#### (Optional) Run as service

```bash
nano redt-agent.service
```

```bash
sudo cp -a ./bin/redt-agent /usr/local/bin/redt-agent
sudo cp -a redt-agent.service /etc/systemd/system/redt-agent.service
sudo systemctl enable redt-agent
sudo systemctl start redt-agent
sudo systemctl status redt-agent
sudo journalctl -u redt-agent

```

### Configuration

Create a configuration file named config.yml in the project directory, and populate it with the following example configuration:

```yaml
backendURL: "https://redt.top/api"
pollInterval: 60 # seconds
upgradeCheckPeriod: 5 # minutes
```

Update the URLs and intervals according to your backend service and requirements.

### Running the Agent

Execute the compiled binary to run the redt-agent:

```bash
./redt-agent
```

The agent will start collecting and reporting data to the backend service based on the configuration file.

### Contributing

Please read CONTRIBUTING.md for details on our code of conduct and the process for submitting pull requests.

### License

This project is licensed under the MIT License - see the LICENSE.md file for details.

### Acknowledgments

The team and contributors who maintain the Go programming language
Everyone who has provided feedback and suggestions for this project

## ⛏️ Built Using <a name = "built_using"></a>

- [MongoDB](https://www.mongodb.com/) - Database
- [Express](https://expressjs.com/) - Server Framework
- [VueJs](https://vuejs.org/) - Web Framework
- [NodeJs](https://nodejs.org/en/) - Server Environment

## ✍️ Authors <a name = "authors"></a>

- [@bluet](https://github.com/bluet) - Idea & Initial work

See also the list of [contributors](https://github.com/bluet/redt-agent/contributors) who participated in this project.

## 🎉 Acknowledgements <a name = "acknowledgement"></a>

- Hat tip to anyone whose code was used
- Inspiration
- References
