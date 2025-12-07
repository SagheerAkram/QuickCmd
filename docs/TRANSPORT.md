# QuickCMD Transport Protocol

## Overview

This document defines the transport protocol for communication between QuickCMD controllers and agents. The protocol uses HTTPS for job submission and WebSocket for real-time log streaming.

## Authentication

### HMAC Signature

All job submissions must include an HMAC-SHA256 signature to verify authenticity and integrity.

**Algorithm:** HMAC-SHA256  
**Secret:** Shared between controller and agent  
**Input:** JSON-serialized job payload  

**Signature Generation:**

```go
// 1. Serialize payload to JSON
data := json.Marshal(payload)

// 2. Create HMAC
h := hmac.New(sha256.New, []byte(secret))
h.Write(data)

// 3. Encode as hex
signature := hex.EncodeToString(h.Sum(nil))
```

**Signature Validation:**

```go
// 1. Generate expected signature from payload
expected := SignPayload(payload, secret)

// 2. Compare using constant-time comparison
valid := hmac.Equal([]byte(expected), []byte(received))
```

### TTL Validation

Jobs include a TTL (Time To Live) to prevent replay attacks.

**Rules:**
- `TTL` must be in the future (Unix timestamp)
- `Timestamp` must be within 5 minutes of current time
- Both checks must pass for job to be accepted

**Example:**
```json
{
  "ttl": 1704628800,        // Must be > current time
  "timestamp": 1704628500   // Must be within 5 minutes of now
}
```

## Job Payload Schema

### SignedJob

Complete job submission including payload and signature.

```json
{
  "payload": {
    "job_id": "string",
    "prompt": "string",
    "command": "string",
    "candidate_metadata": {},
    "plugin_metadata": {},
    "required_scopes": ["string"],
    "snapshot_metadata": "string",
    "ttl": 1704628800,
    "timestamp": 1704628500,
    "controller_id": "string"
  },
  "signature": {
    "signature": "hex-encoded-hmac-sha256",
    "algorithm": "HMAC-SHA256"
  }
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `job_id` | string | Yes | Unique identifier for the job |
| `prompt` | string | Yes | Original natural language prompt |
| `command` | string | Yes | Command to execute |
| `candidate_metadata` | object | No | Metadata from candidate selection |
| `plugin_metadata` | object | No | Plugin-specific metadata |
| `required_scopes` | array | No | Required permission scopes |
| `snapshot_metadata` | string | No | JSON-encoded snapshot info |
| `ttl` | integer | Yes | Unix timestamp when job expires |
| `timestamp` | integer | Yes | Unix timestamp when job was created |
| `controller_id` | string | Yes | Identifier of submitting controller |

## HTTP Endpoints

### Submit Job

**Endpoint:** `POST /api/v1/jobs`  
**Content-Type:** `application/json`  
**Authentication:** HMAC signature in payload  

**Request:**
```json
{
  "payload": { /* JobPayload */ },
  "signature": {
    "signature": "abc123...",
    "algorithm": "HMAC-SHA256"
  }
}
```

**Success Response (202 Accepted):**
```json
{
  "job_id": "job-123",
  "status": "pending"
}
```

**Error Responses:**

| Status | Error | Description |
|--------|-------|-------------|
| 400 | Invalid JSON | Malformed request body |
| 401 | Invalid signature | HMAC validation failed |
| 401 | Job expired | TTL exceeded |
| 401 | Job too old | Timestamp too old (replay attack) |
| 403 | Controller not allowed | Controller ID not in allowlist |
| 500 | Internal error | Server error |

### Get Job Status

**Endpoint:** `GET /api/v1/jobs/:id`  
**Authentication:** None (job ID acts as token)  

**Success Response (200 OK):**
```json
{
  "job_id": "job-123",
  "status": "completed",
  "result": {
    "job_id": "job-123",
    "status": "completed",
    "sandbox_id": "abc123",
    "exit_code": 0,
    "stdout": "command output",
    "stderr": "",
    "start_time": "2025-01-07T10:00:00Z",
    "end_time": "2025-01-07T10:00:05Z",
    "duration_ms": 5000,
    "error": "",
    "snapshot": ""
  }
}
```

**Job Statuses:**
- `pending` - Job accepted, waiting to execute
- `running` - Job currently executing
- `completed` - Job finished successfully
- `failed` - Job failed with error
- `rejected` - Job rejected by policy/plugin checks

## WebSocket Log Streaming

### Connection

**Endpoint:** `wss://agent:8443/api/v1/stream/:id`  
**Protocol:** WebSocket  
**Authentication:** Origin check against allowed controllers  

### Log Frame Format

Each log message is sent as a JSON frame:

```json
{
  "job_id": "job-123",
  "timestamp": "2025-01-07T10:00:01.123Z",
  "stream": "stdout",
  "data": "Log message content",
  "final": false
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `job_id` | string | Job identifier |
| `timestamp` | string | ISO 8601 timestamp |
| `stream` | string | Either "stdout" or "stderr" |
| `data` | string | Log message content |
| `final` | boolean | True for the last frame |

### Frame Sequence

1. **Initial frames:** `final: false`, contain log data
2. **Intermediate frames:** `final: false`, contain log data
3. **Final frame:** `final: true`, signals end of stream

**Example Sequence:**

```json
{"job_id":"job-123","timestamp":"...","stream":"stdout","data":"Starting...","final":false}
{"job_id":"job-123","timestamp":"...","stream":"stdout","data":"Processing...","final":false}
{"job_id":"job-123","timestamp":"...","stream":"stdout","data":"Complete","final":false}
{"job_id":"job-123","timestamp":"...","stream":"stdout","data":"","final":true}
```

### Connection Lifecycle

```
Client                                Agent
  │                                     │
  ├─── WebSocket Upgrade ──────────────▶│
  │◀──── 101 Switching Protocols ───────┤
  │                                     │
  │◀──── Log Frame (final: false) ──────┤
  │◀──── Log Frame (final: false) ──────┤
  │◀──── Log Frame (final: false) ──────┤
  │◀──── Log Frame (final: true) ───────┤
  │                                     │
  ├─── Close ───────────────────────────▶│
  │◀──── Close ──────────────────────────┤
  │                                     │
```

## Error Handling

### Transient Errors

Errors that may be resolved by retrying:
- Network timeouts
- Connection refused
- 5xx server errors

**Retry Strategy:**
- Exponential backoff: 1s, 2s, 4s
- Maximum 3 retries
- Give up on 4xx errors (client errors)

**Example:**
```go
for attempt := 0; attempt <= maxRetries; attempt++ {
    if attempt > 0 {
        backoff := time.Duration(1<<uint(attempt-1)) * time.Second
        time.Sleep(backoff)
    }
    
    resp, err := submitJob(job)
    if err == nil && resp.StatusCode == 202 {
        return resp
    }
    
    // Don't retry on client errors
    if resp.StatusCode >= 400 && resp.StatusCode < 500 {
        return err
    }
}
```

### Permanent Errors

Errors that should not be retried:
- 401 Unauthorized (invalid signature)
- 403 Forbidden (controller not allowed)
- 400 Bad Request (malformed payload)

## Security Considerations

### Signature Validation

**CRITICAL:** Always use constant-time comparison for signatures to prevent timing attacks.

```go
// ✓ Correct
hmac.Equal([]byte(expected), []byte(received))

// ✗ Wrong (vulnerable to timing attacks)
expected == received
```

### TTL Best Practices

- Set TTL to minimum necessary (e.g., 5 minutes)
- Validate both TTL and timestamp
- Reject jobs with timestamps too far in past

### TLS Configuration

**Production Requirements:**
- Use TLS 1.2 or higher
- Use strong cipher suites
- Validate server certificates
- Use proper CA-signed certificates

**Development:**
- Self-signed certificates acceptable
- Set `InsecureSkipVerify: true` only for development

### Secret Management

**NEVER:**
- Commit HMAC secrets to version control
- Send secrets in plain text
- Reuse secrets across environments

**DO:**
- Generate cryptographically random secrets
- Rotate secrets regularly
- Use different secrets per agent
- Store secrets in secure key management systems

## Example Implementations

### Controller: Submit Job

```go
package main

import (
    "context"
    "time"
    "github.com/yourusername/quickcmd/agent"
    "github.com/yourusername/quickcmd/controller"
)

func main() {
    // Create client
    client := controller.NewClient("https://agent.example.com:8443", "hmac-secret")
    
    // Create job payload
    payload := &agent.JobPayload{
        JobID:        "job-123",
        Prompt:       "list files",
        Command:      "ls -la",
        TTL:          time.Now().Add(5 * time.Minute).Unix(),
        Timestamp:    time.Now().Unix(),
        ControllerID: "controller-1",
    }
    
    // Submit job
    ctx := context.Background()
    jobID, err := client.SubmitJob(ctx, payload)
    if err != nil {
        panic(err)
    }
    
    println("Job submitted:", jobID)
}
```

### Controller: Stream Logs

```go
// Stream logs from job
err := client.StreamLogs(ctx, jobID, func(frame *agent.LogFrame) error {
    fmt.Printf("[%s] %s: %s\n", frame.Timestamp, frame.Stream, frame.Data)
    return nil
})
```

### Controller: Wait for Completion

```go
// Wait for job to complete
result, err := client.WaitForCompletion(ctx, jobID, 2*time.Second)
if err != nil {
    panic(err)
}

fmt.Printf("Exit code: %d\n", result.ExitCode)
fmt.Printf("Output: %s\n", result.Stdout)
```

## Monitoring

### Metrics

Agents expose Prometheus metrics at `/metrics`:

```
quickcmd_agent_jobs_total 42
quickcmd_agent_jobs_running 2
quickcmd_agent_jobs_completed 38
quickcmd_agent_jobs_failed 2
```

### Health Checks

Controllers should poll `/health` endpoint:

```bash
curl https://agent.example.com:8443/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": 1704628500
}
```

## Versioning

Current API version: **v1**

API endpoint format: `/api/v1/...`

**Compatibility Promise:**
- v1 API will remain stable
- Breaking changes will use v2 endpoints
- Deprecation warnings provided 6 months before removal

---

**For implementation details, see AGENT.md and controller source code.**
