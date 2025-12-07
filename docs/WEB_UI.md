# QuickCMD Web UI

## Overview

The QuickCMD Web UI provides a team-friendly interface for reviewing execution history, managing approvals, and monitoring command execution across distributed agents.

## Features

- **Authentication**: JWT-based auth with RBAC (viewer/operator/approver/admin)
- **History View**: Paginated execution history with filtering
- **Run Details**: Complete metadata, outputs, and snapshot information
- **Approval Workflow**: Review and approve/reject pending jobs with typed confirmation
- **Security**: CORS, CSRF protection, secrets redaction

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker (for agent execution)

### Installation

```bash
# Install backend dependencies
cd web
go mod download

# Install frontend dependencies
cd frontend
npm install
```

### Development Mode

**1. Start Backend API:**

```bash
# From web/ directory
go run cmd/web-server/main.go
```

Backend runs on `http://localhost:3000`

**2. Start Frontend:**

```bash
# From web/frontend/ directory
npm run dev
```

Frontend runs on `http://localhost:3001`

**3. Login:**

Navigate to `http://localhost:3001/login`

**Dev Mode Users:**
- `admin` / `admin` - Full access
- `approver` / `approver` - Can approve jobs
- `operator` / `operator` - Can execute
- `viewer` / `viewer` - Read-only

## Configuration

### Environment Variables

```bash
# Backend
export WEB_PORT=3000
export WEB_ORIGINS="http://localhost:3001,http://localhost:3000"
export JWT_SECRET="your-secret-key-here"
export AUDIT_DB_PATH="/var/lib/quickcmd/audit.db"
export APPROVAL_DB_PATH="/var/lib/quickcmd/approvals.db"

# Frontend
export NEXT_PUBLIC_API_URL="http://localhost:3000"
```

### Production Configuration

**Backend (`web-config.yaml`):**

```yaml
port: 3000
cors_origins:
  - "https://quickcmd.example.com"
auth:
  jwt_secret: "CHANGE_ME_IN_PRODUCTION"
  token_duration: "15m"
  dev_mode: false
  oidc_enabled: true
  oidc:
    issuer: "https://auth.example.com"
    client_id: "quickcmd"
    client_secret: "secret"
    redirect_url: "https://quickcmd.example.com/auth/callback"
audit_db_path: "/var/lib/quickcmd/audit.db"
approval_db_path: "/var/lib/quickcmd/approvals.db"
```

## API Endpoints

### Authentication

**POST /api/v1/login**

Login with username/password, returns JWT token.

```bash
curl -X POST http://localhost:3000/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900
}
```

### History

**GET /api/v1/history**

Get execution history with optional filtering.

```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:3000/api/v1/history?limit=20&filter=docker"
```

### Run Details

**GET /api/v1/run/:id**

Get detailed information about a specific run.

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:3000/api/v1/run/123
```

### Approvals

**GET /api/v1/approvals**

List pending approvals.

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:3000/api/v1/approvals
```

**POST /api/v1/approvals/:id/approve**

Approve a pending job (requires approver role).

```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"confirmation":"APPROVE 123","note":"Approved for deployment"}' \
  http://localhost:3000/api/v1/approvals/123/approve
```

**POST /api/v1/approvals/:id/reject**

Reject a pending job.

```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"reason":"Insufficient testing"}' \
  http://localhost:3000/api/v1/approvals/123/reject
```

## Security

### CORS Configuration

Configure allowed origins in `WEB_ORIGINS` environment variable:

```bash
export WEB_ORIGINS="https://app.example.com,https://admin.example.com"
```

### CSRF Protection

All state-changing endpoints require CSRF token:

```javascript
// Get CSRF token
const response = await fetch('/api/v1/csrf-token');
const { csrf_token } = await response.json();

// Include in requests
fetch('/api/v1/approvals/123/approve', {
  method: 'POST',
  headers: {
    'X-CSRF-Token': csrf_token,
    'Authorization': `Bearer ${token}`,
  },
  body: JSON.stringify({ confirmation: 'APPROVE 123' }),
});
```

### Role-Based Access Control

| Role | Permissions |
|------|-------------|
| **viewer** | View history, run details |
| **operator** | viewer + execute commands |
| **approver** | operator + approve/reject jobs |
| **admin** | All permissions + user management |

### Secrets Redaction

All outputs are automatically redacted using the existing redaction module:
- Environment variables (PASSWORD, TOKEN, API_KEY, SECRET)
- Bearer tokens
- AWS credentials
- Private keys
- Database connection strings

## Troubleshooting

### Backend Won't Start

**Error: "Failed to open database"**

```bash
# Create database directory
sudo mkdir -p /var/lib/quickcmd
sudo chown $USER /var/lib/quickcmd
```

### Frontend Can't Connect to Backend

**Error: "CORS policy"**

Ensure backend `WEB_ORIGINS` includes frontend URL:

```bash
export WEB_ORIGINS="http://localhost:3001"
```

### Authentication Fails

**Error: "Invalid or expired token"**

- Check JWT_SECRET matches between sessions
- Token expires after 15 minutes by default
- Re-login to get new token

## Production Deployment

### Backend

```bash
# Build
go build -o quickcmd-web cmd/web-server/main.go

# Run with systemd
sudo cp quickcmd-web /usr/local/bin/
sudo cp deployments/systemd/quickcmd-web.service /etc/systemd/system/
sudo systemctl start quickcmd-web
sudo systemctl enable quickcmd-web
```

### Frontend

```bash
# Build
npm run build

# Start
npm start

# Or use PM2
pm2 start npm --name "quickcmd-web" -- start
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl;
    server_name quickcmd.example.com;

    ssl_certificate /etc/letsencrypt/live/quickcmd.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/quickcmd.example.com/privkey.pem;

    # Frontend
    location / {
        proxy_pass http://localhost:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api/ {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## OIDC Integration (Optional)

### Configuration

```yaml
auth:
  oidc_enabled: true
  oidc:
    issuer: "https://accounts.google.com"
    client_id: "your-client-id"
    client_secret: "your-client-secret"
    redirect_url: "https://quickcmd.example.com/auth/callback"
```

### Role Mapping

Map OIDC groups to QuickCMD roles:

```yaml
role_mapping:
  "quickcmd-admins": "admin"
  "quickcmd-approvers": "approver"
  "quickcmd-operators": "operator"
  "quickcmd-viewers": "viewer"
```

---

**For approval workflow details, see [APPROVALS.md](./APPROVALS.md)**
