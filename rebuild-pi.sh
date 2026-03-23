#!/bin/bash
set -e

echo "🏟  Arena — Raspberry Pi Docker Rebuild Script"
echo "=============================================="

echo "⬇️  Pulling latest changes from git..."
git pull

echo "📦 Building Docker image (arena-monitor)..."
docker build -t arena-monitor .

echo "🛑 Stopping and removing existing container (if any)..."
docker stop arena 2>/dev/null || true
docker rm arena 2>/dev/null || true

echo "🚀 Starting new 'arena' container..."
# Using host networking is REQUIRED for Nmap ARP sweeps on Linux/Pi
docker run -d \
  --name arena \
  --network host \
  --restart unless-stopped \
  -v "$(pwd)/devices.json:/app/devices.json" \
  -v "$(pwd)/config.yaml:/app/config.yaml" \
  arena-monitor

echo "✅ Deployment complete! The Arena daemon should be running on the host network."
echo "   Monitor logs using: docker logs -f arena"
