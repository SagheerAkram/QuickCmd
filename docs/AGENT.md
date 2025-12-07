# QuickCMD Remote Agent

## Overview

The QuickCMD Remote Agent allows secure, distributed command execution across multiple hosts. Controllers submit signed jobs to agents, which execute commands in isolated Docker sandboxes and stream results back in real-time.

## Architecture

```
┌─────────────┐                    ┌──────────────┐
│ Controller  │ ──── HTTPS ────────▶│    Agent     │
│             │ ◀─── WebSocket ─────│              │
└─────────────┘                    └──────────────┘
      │                                    │
      │ Sign Job (HMAC)                   │ Validate
      │ Submit                             │ Execute in Sandbox
      │ Stream Logs                        │ Stream Logs
      │ Get Result                         │ Audit Log
      │                                    │
      ▼                                    ▼
┌─────────────┐                    ┌──────────────┐
│  Audit DB   │                    │  Docker      │
└─────────────┘                    └──────────────┘
```

## Installation

### Prerequisites

- Go 1.21+
- Docker (for sandbox execution)
- Linux system (Debian/Ubuntu recommended)
- OpenSSL (for TLS certificates)

### Quick Install

```bash
# Clone repository
git clone https://github.com/yourusername/quickcmd
cd quickcmd

# Build agent
make build-agent

# Run setup script (as root)
sudo ./scripts/setup-agent.sh
```

The setup script will:
1. Create `quickcmd` user and group
2. Create directories (`/etc/quickcmd`, `/var/lib/quickcmd`)
3. Install binary to `/usr/local/bin/quickcmd-agent`
4. Generate self-signed TLS certificates
5. Install systemd service
6. Add `quickcmd` user to `docker` group

### Manual Installation

```bash
# Create user
sudo useradd --system --no-create-home --shell /bin/false quickcmd

# Create directories
sudo mkdir -p /etc/quickcmd /var/lib/quickcmd
sudo chown quickcmd:quickcmd /var/lib/quickcmd

# Copy binary
sudo cp bin/quickcmd-agent /usr/local/bin/
sudo chmod 755 /usr/local/bin/quickcmd-agent

# Copy configuration
sudo cp examples/agent-config.yaml /etc/quickcmd/
sudo chmod 600 /etc/quickcmd/agent-config.yaml
sudo chown quickcmd:quickcmd /etc/quickcmd/agent-config.yaml

# Generate TLS certificates
sudo openssl req -x509 -newkey rsa:4096 \
  -keyout /etc/quickcmd/agent-key.pem \
  -out /etc/quickcmd/agent-cert.pem \
  -days 365 -nodes \
  -subj "/CN=quickcmd-agent"

# Install systemd service
sudo cp deployments/systemd/quickcmd-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
```

## Configuration

Edit `/etc/quickcmd/agent-config.yaml`:

```yaml
# Server settings
port: 8443
tls_cert_file: "/etc/quickcmd/agent-cert.pem"
tls_key_file: "/etc/quickcmd/agent-key.pem"

# Authentication
hmac_secret: "CHANGE_ME"  # Generate with: quickcmd agent gen-key

# Allowed controllers
allowed_controllers:
  - "controller-1"
  - "https://quickcmd.example.com"

# Execution settings
max_concurrent_jobs: 5
allowed_images:
  - "alpine:latest"
  - "ubuntu:latest"
default_image: "alpine:latest"

# Security
run_as_user: "quickcmd"
run_as_group: "quickcmd"

# Sandbox defaults
default_cpu_limit: 0.5
default_memory_limit: 268435456  # 256 MB
default_timeout_seconds: 300

# Audit
audit_db_path: "/var/lib/quickcmd/agent-audit.db"
```

### Generate HMAC Secret

```bash
# Generate a secure random secret
openssl rand -hex 32
```

Update `hmac_secret` in the configuration file with the generated value.

## Running the Agent

### Using Systemd

```bash
# Start agent
sudo systemctl start quickcmd-agent

# Enable on boot
sudo systemctl enable quickcmd-agent

# Check status
sudo systemctl status quickcmd-agent

# View logs
sudo journalctl -u quickcmd-agent -f
```

### Manual Run (Development)

```bash
quickcmd-agent --config=/etc/quickcmd/agent-config.yaml
```

## API Endpoints

### Submit Job

**POST** `/api/v1/jobs`

Submit a signed job for execution.

**Request:**
```json
{
  "payload": {
    "job_id": "job-123",
    "prompt": "list files",
    "command": "ls -la",
    "candidate_metadata": {},
    "plugin_metadata": {},
    "required_scopes": [],
    "snapshot_metadata": "",
    "ttl": 1704628800,
    "timestamp": 1704628500,
    "controller_id": "controller-1"
  },
  "signature": {
    "signature": "abc123...",
    "algorithm": "HMAC-SHA256"
  }
}
```

**Response:**
```json
{
  "job_id": "job-123",
  "status": "pending"
}
```

### Get Job Status

**GET** `/api/v1/jobs/:id`

Get the current status and result of a job.

**Response:**
```json
{
  "job_id": "job-123",
  "status": "completed",
  "result": {
    "job_id": "job-123",
    "status": "completed",
    "sandbox_id": "abc123",
    "exit_code": 0,
    "stdout": "...",
    "stderr": "",
    "start_time": "2025-01-07T10:00:00Z",
    "end_time": "2025-01-07T10:00:05Z",
    "duration_ms": 5000
  }
}
```

### Stream Logs

**WebSocket** `/api/v1/stream/:id`

Stream real-time logs from a running job.

**Log Frame:**
```json
{
  "job_id": "job-123",
  "timestamp": "2025-01-07T10:00:01Z",
  "stream": "stdout",
  "data": "Starting execution...",
  "final": false
}
```

### Health Check

**GET** `/health`

Returns agent health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1704628500
}
```

### Metrics

**GET** `/metrics`

Returns Prometheus metrics.

```
# HELP quickcmd_agent_jobs_total Total number of jobs
# TYPE quickcmd_agent_jobs_total counter
quickcmd_agent_jobs_total 42

# HELP quickcmd_agent_jobs_running Currently running jobs
# TYPE quickcmd_agent_jobs_running gauge
quickcmd_agent_jobs_running 2
```

## Security

### TLS Configuration

**Production:** Use proper TLS certificates from a trusted CA.

```bash
# Example with Let's Encrypt
sudo certbot certonly --standalone -d agent.example.com
```

Update configuration:
```yaml
tls_cert_file: "/etc/letsencrypt/live/agent.example.com/fullchain.pem"
tls_key_file: "/etc/letsencrypt/live/agent.example.com/privkey.pem"
```

### HMAC Secret Management

- **Never commit secrets to version control**
- Use environment variables or secret management systems
- Rotate secrets regularly
- Use different secrets for each agent

### Firewall Configuration

```bash
# Allow only specific controller IPs
sudo ufw allow from 192.168.1.100 to any port 8443
sudo ufw enable
```

### Systemd Security

The provided systemd unit includes security hardening:
- `NoNewPrivileges=true` - Prevents privilege escalation
- `PrivateTmp=true` - Isolated /tmp
- `ProtectSystem=strict` - Read-only system directories
- `ProtectHome=true` - No access to home directories
- `RestrictNamespaces=true` - Limited namespace access
- `SystemCallFilter=@system-service` - Syscall filtering

## Monitoring

### Prometheus Integration

Add to Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'quickcmd-agent'
    static_configs:
      - targets: ['agent.example.com:8443']
    scheme: https
    tls_config:
      insecure_skip_verify: true  # Remove in production
```

### Log Aggregation

Agent logs to systemd journal. Forward to centralized logging:

```bash
# Example with journalbeat
sudo apt install journalbeat
```

## Troubleshooting

### Agent Won't Start

**Check logs:**
```bash
sudo journalctl -u quickcmd-agent -n 50
```

**Common issues:**
- Docker not running: `sudo systemctl start docker`
- Permission denied: Check file permissions on config and certs
- Port in use: Change port in configuration

### Job Submission Fails

**Error: Invalid signature**
- Verify HMAC secret matches between controller and agent
- Check job TTL hasn't expired
- Verify timestamp is current

**Error: Controller not allowed**
- Add controller ID to `allowed_controllers` in config
- Restart agent after config changes

### Docker Errors

**Error: Cannot connect to Docker daemon**
```bash
# Add quickcmd user to docker group
sudo usermod -aG docker quickcmd

# Restart agent
sudo systemctl restart quickcmd-agent
```

## Upgrading

```bash
# Stop agent
sudo systemctl stop quickcmd-agent

# Backup configuration and audit database
sudo cp /etc/quickcmd/agent-config.yaml /etc/quickcmd/agent-config.yaml.backup
sudo cp /var/lib/quickcmd/agent-audit.db /var/lib/quickcmd/agent-audit.db.backup

# Install new binary
sudo cp bin/quickcmd-agent /usr/local/bin/

# Start agent
sudo systemctl start quickcmd-agent
```

## Uninstallation

```bash
# Stop and disable service
sudo systemctl stop quickcmd-agent
sudo systemctl disable quickcmd-agent

# Remove files
sudo rm /usr/local/bin/quickcmd-agent
sudo rm /etc/systemd/system/quickcmd-agent.service
sudo rm -rf /etc/quickcmd
sudo rm -rf /var/lib/quickcmd

# Remove user
sudo userdel quickcmd

# Reload systemd
sudo systemctl daemon-reload
```

---

**For controller integration and job submission examples, see the main QuickCMD documentation.**
