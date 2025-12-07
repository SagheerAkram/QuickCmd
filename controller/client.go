package controller

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"github.com/gorilla/websocket"
	"github.com/SagheerAkram/QuickCmd/agent"
)

// Client represents a controller client for submitting jobs to agents
type Client struct {
	agentURL   string
	hmacSecret string
	httpClient *http.Client
	maxRetries int
}

// NewClient creates a new controller client
func NewClient(agentURL, hmacSecret string) *Client {
	return &Client{
		agentURL:   agentURL,
		hmacSecret: hmacSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // For development with self-signed certs
				},
			},
		},
		maxRetries: 3,
	}
}

// SubmitJob submits a job to the agent with retry logic
func (c *Client) SubmitJob(ctx context.Context, payload *agent.JobPayload) (string, error) {
	// Sign the payload
	signature, err := agent.SignPayload(payload, c.hmacSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign payload: %w", err)
	}
	
	signedJob := &agent.SignedJob{
		Payload: *payload,
		Signature: agent.JobSignature{
			Signature: signature,
			Algorithm: "HMAC-SHA256",
		},
	}
	
	// Serialize to JSON
	data, err := json.Marshal(signedJob)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}
	
	// Submit with retry
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}
		
		req, err := http.NewRequestWithContext(ctx, "POST", c.agentURL+"/api/v1/jobs", bytes.NewReader(data))
		if err != nil {
			lastErr = err
			continue
		}
		
		req.Header.Set("Content-Type", "application/json")
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		
		if resp.StatusCode == http.StatusAccepted {
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				lastErr = err
				continue
			}
			resp.Body.Close()
			
			jobID, _ := result["job_id"].(string)
			return jobID, nil
		}
		
		// Read error response
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastErr = fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
		
		// Don't retry on client errors
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return "", lastErr
		}
	}
	
	return "", fmt.Errorf("failed after %d retries: %w", c.maxRetries, lastErr)
}

// GetJobStatus retrieves the status of a job
func (c *Client) GetJobStatus(ctx context.Context, jobID string) (*agent.JobResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.agentURL+"/api/v1/jobs/"+jobID, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		JobID  string            `json:"job_id"`
		Status agent.JobStatus   `json:"status"`
		Result *agent.JobResult  `json:"result"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	return response.Result, nil
}

// StreamLogs streams logs from a job via WebSocket
func (c *Client) StreamLogs(ctx context.Context, jobID string, logHandler func(*agent.LogFrame) error) error {
	// Create WebSocket connection
	wsURL := "wss://" + c.agentURL[8:] + "/api/v1/stream/" + jobID // Replace https:// with wss://
	
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // For development
		},
	}
	
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()
	
	// Read log frames
	for {
		var frame agent.LogFrame
		if err := conn.ReadJSON(&frame); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil
			}
			return fmt.Errorf("failed to read log frame: %w", err)
		}
		
		if err := logHandler(&frame); err != nil {
			return err
		}
		
		if frame.Final {
			return nil
		}
	}
}

// WaitForCompletion waits for a job to complete and returns the result
func (c *Client) WaitForCompletion(ctx context.Context, jobID string, pollInterval time.Duration) (*agent.JobResult, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			result, err := c.GetJobStatus(ctx, jobID)
			if err != nil {
				return nil, err
			}
			
			if result != nil && (result.Status == agent.JobStatusCompleted || result.Status == agent.JobStatusFailed) {
				return result, nil
			}
		}
	}
}
