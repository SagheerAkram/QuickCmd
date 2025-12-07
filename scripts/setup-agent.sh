#!/bin/bash
# QuickCMD Agent Setup Script for Debian/Ubuntu

set -e

echo "QuickCMD Agent Setup"
echo "===================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

# Create quickcmd user and group
echo "Creating quickcmd user and group..."
if ! id -u quickcmd > /dev/null 2>&1; then
  useradd --system --no-create-home --shell /bin/false quickcmd
fi

# Create directories
echo "Creating directories..."
mkdir -p /etc/quickcmd
mkdir -p /var/lib/quickcmd
mkdir -p /var/log/quickcmd

# Set permissions
chown quickcmd:quickcmd /var/lib/quickcmd
chown quickcmd:quickcmd /var/log/quickcmd
chmod 750 /etc/quickcmd
chmod 750 /var/lib/quickcmd

# Copy binary
echo "Installing binary..."
if [ -f "./bin/quickcmd-agent" ]; then
  cp ./bin/quickcmd-agent /usr/local/bin/
  chmod 755 /usr/local/bin/quickcmd-agent
else
  echo "Error: Binary not found at ./bin/quickcmd-agent"
  echo "Please build the binary first: make build-agent"
  exit 1
fi

# Copy configuration
echo "Installing configuration..."
if [ ! -f "/etc/quickcmd/agent-config.yaml" ]; then
  if [ -f "./examples/agent-config.yaml" ]; then
    cp ./examples/agent-config.yaml /etc/quickcmd/
    chmod 600 /etc/quickcmd/agent-config.yaml
    chown quickcmd:quickcmd /etc/quickcmd/agent-config.yaml
    echo "WARNING: Please edit /etc/quickcmd/agent-config.yaml and set your HMAC secret!"
  else
    echo "Error: Example config not found"
    exit 1
  fi
else
  echo "Configuration already exists at /etc/quickcmd/agent-config.yaml"
fi

# Install systemd service
echo "Installing systemd service..."
if [ -f "./deployments/systemd/quickcmd-agent.service" ]; then
  cp ./deployments/systemd/quickcmd-agent.service /etc/systemd/system/
  systemctl daemon-reload
else
  echo "Error: Systemd unit file not found"
  exit 1
fi

# Generate TLS certificates (self-signed for development)
echo "Generating self-signed TLS certificates..."
if [ ! -f "/etc/quickcmd/agent-cert.pem" ]; then
  openssl req -x509 -newkey rsa:4096 -keyout /etc/quickcmd/agent-key.pem \
    -out /etc/quickcmd/agent-cert.pem -days 365 -nodes \
    -subj "/CN=quickcmd-agent/O=QuickCMD/C=US"
  chmod 600 /etc/quickcmd/agent-key.pem
  chmod 644 /etc/quickcmd/agent-cert.pem
  chown quickcmd:quickcmd /etc/quickcmd/agent-*.pem
  echo "Self-signed certificates generated. For production, use proper certificates."
else
  echo "TLS certificates already exist"
fi

# Add quickcmd user to docker group
echo "Adding quickcmd user to docker group..."
usermod -aG docker quickcmd

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit /etc/quickcmd/agent-config.yaml and set your HMAC secret"
echo "2. Update allowed_controllers list"
echo "3. Start the service: systemctl start quickcmd-agent"
echo "4. Enable on boot: systemctl enable quickcmd-agent"
echo "5. Check status: systemctl status quickcmd-agent"
echo "6. View logs: journalctl -u quickcmd-agent -f"
echo ""
echo "For production, replace self-signed certificates with proper TLS certificates."
