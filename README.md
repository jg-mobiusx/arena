# 🏟 Arena — Network Monitor

Arena is a lightweight network topology scanner and React dashboard designed primarily for Raspberry Pi deployments. It automatically discovers, enriches, and visualizes network hardware using raw ARP sweeps.

## Architecture

Arena is composed of two primary layers:
1. **Go Daemon (`arena`)**: A modular backend that leverages `nmap` for precise ARP sweeps. It discovers active IPs/MACs. It saves network state recursively to a persistent `devices.json` store and acts as an embedded web server.
2. **React Dashboard (`web`)**: A powerful Vite/React frontend using `@xyflow/react` to render complex network topographies segmented by VLAN buses.

## Requirements
- Go 1.22+
- Node.js 20+
- `nmap` installed on the host machine (requires `root` privileges to scan)


## Running Locally

**Start the React Development Server:**
```bash
cd web
npm install
npm run dev
```

**Start the Go Daemon (with privileges):**
Nmap requires root access for Layer 2 ARP sweeps. The Go binary spawns `nmap`, so it must be run with elevated privileges.
```bash
sudo go run main.go
```

By default, the daemon runs continuously every 5 minutes and serves the dashboard on **http://localhost:8080**.

## Deployment (Docker)

A multi-stage `Dockerfile` is included to compile the Node frontend, compile the Go binary, and deploy them together into a lightweight Alpine container with Nmap. 

Because Nmap needs raw Layer 2 socket access to discover MAC addresses on the local network, you must deploy the container using **host networking**. 

```bash
docker build -t arena-monitor .
docker run -d --name arena --network host arena-monitor
```
*(Note: Host networking is natively supported on Linux mechanisms like the Raspberry Pi, but acts as a virtual NAT on Docker Desktop for macOS, preventing proper ARP sweeps. For Mac development, run the binary directly via Go).*

## Configuration
Edit `config.yaml` to specify which subnets to scan.
