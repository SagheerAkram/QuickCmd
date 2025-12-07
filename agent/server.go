package agent

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	
	"github.com/gorilla/websocket"
)

// Server represents the agent HTTP server
type Server struct {
	config    *Config
	jobs      map[string]*Job
	jobsMu    sync.RWMutex
	upgrader  websocket.Upgrader
	executor  *JobExecutor
	httpServer *http.Server
}

// Job represents a job being executed
type Job struct {
	Payload   *JobPayload
	Status    JobStatus
	Result    *JobResult
	LogChan   chan *LogFrame
	CancelFunc context.CancelFunc
	CreatedAt time.Time
}

// NewServer creates a new agent server
func NewServer(config *Config) (*Server, error) {
	executor, err := NewJobExecutor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}
	
	server := &Server{
		config:   config,
		jobs:     make(map[string]*Job),
		executor: executor,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Check if origin is in allowed controllers
				origin := r.Header.Get("Origin")
				for _, allowed := range config.AllowedControllers {
					if origin == allowed {
						return true
					}
				}
				return false
			},
		},
	}
	
	return server, nil
}

// Start starts the HTTPS server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// Register handlers
	mux.HandleFunc("/api/v1/jobs", s.handleSubmitJob)
	mux.HandleFunc("/api/v1/jobs/", s.handleJobStatus)
	mux.HandleFunc("/api/v1/stream/", s.handleLogStream)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)
	
	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	log.Printf("Starting agent server on port %d", s.config.Port)
	
	if s.config.TLSCertFile != "" && s.config.TLSKeyFile != "" {
		return s.httpServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	}
	
	log.Println("WARNING: Running without TLS (development mode only)")
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Cancel all running jobs
	s.jobsMu.Lock()
	for _, job := range s.jobs {
		if job.CancelFunc != nil {
			job.CancelFunc()
		}
	}
	s.jobsMu.Unlock()
	
	return s.httpServer.Shutdown(ctx)
}

// handleSubmitJob handles job submission
func (s *Server) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse signed job
	var signedJob SignedJob
	if err := json.NewDecoder(r.Body).Decode(&signedJob); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}
	
	// Validate signature
	if err := ValidateSignature(&signedJob, s.config.HMACSecret); err != nil {
		s.writeError(w, http.StatusUnauthorized, "Invalid signature", err)
		return
	}
	
	// Validate TTL
	if err := ValidateTTL(&signedJob.Payload); err != nil {
		s.writeError(w, http.StatusUnauthorized, "Job expired or too old", err)
		return
	}
	
	// Validate controller
	if !s.isAllowedController(signedJob.Payload.ControllerID) {
		s.writeError(w, http.StatusForbidden, "Controller not allowed", nil)
		return
	}
	
	// Create job
	ctx, cancel := context.WithCancel(context.Background())
	job := &Job{
		Payload:    &signedJob.Payload,
		Status:     JobStatusPending,
		LogChan:    make(chan *LogFrame, 100),
		CancelFunc: cancel,
		CreatedAt:  time.Now(),
	}
	
	// Store job
	s.jobsMu.Lock()
	s.jobs[signedJob.Payload.JobID] = job
	s.jobsMu.Unlock()
	
	// Execute job asynchronously
	go s.executeJob(ctx, job)
	
	// Return job ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": signedJob.Payload.JobID,
		"status": JobStatusPending,
	})
}

// handleJobStatus returns the status of a job
func (s *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract job ID from path
	jobID := r.URL.Path[len("/api/v1/jobs/"):]
	
	s.jobsMu.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMu.RUnlock()
	
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": jobID,
		"status": job.Status,
		"result": job.Result,
	})
}

// handleLogStream handles WebSocket log streaming
func (s *Server) handleLogStream(w http.ResponseWriter, r *http.Request) {
	// Extract job ID
	jobID := r.URL.Path[len("/api/v1/stream/"):]
	
	s.jobsMu.RLock()
	job, exists := s.jobs[jobID]
	s.jobsMu.RUnlock()
	
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	
	// Upgrade to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	
	// Stream logs
	for frame := range job.LogChan {
		if err := conn.WriteJSON(frame); err != nil {
			log.Printf("Failed to write log frame: %v", err)
			return
		}
		
		if frame.Final {
			break
		}
	}
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// handleMetrics returns Prometheus metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	s.jobsMu.RLock()
	totalJobs := len(s.jobs)
	
	var running, completed, failed int
	for _, job := range s.jobs {
		switch job.Status {
		case JobStatusRunning:
			running++
		case JobStatusCompleted:
			completed++
		case JobStatusFailed:
			failed++
		}
	}
	s.jobsMu.RUnlock()
	
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP quickcmd_agent_jobs_total Total number of jobs\n")
	fmt.Fprintf(w, "# TYPE quickcmd_agent_jobs_total counter\n")
	fmt.Fprintf(w, "quickcmd_agent_jobs_total %d\n", totalJobs)
	
	fmt.Fprintf(w, "# HELP quickcmd_agent_jobs_running Currently running jobs\n")
	fmt.Fprintf(w, "# TYPE quickcmd_agent_jobs_running gauge\n")
	fmt.Fprintf(w, "quickcmd_agent_jobs_running %d\n", running)
	
	fmt.Fprintf(w, "# HELP quickcmd_agent_jobs_completed Completed jobs\n")
	fmt.Fprintf(w, "# TYPE quickcmd_agent_jobs_completed counter\n")
	fmt.Fprintf(w, "quickcmd_agent_jobs_completed %d\n", completed)
	
	fmt.Fprintf(w, "# HELP quickcmd_agent_jobs_failed Failed jobs\n")
	fmt.Fprintf(w, "# TYPE quickcmd_agent_jobs_failed counter\n")
	fmt.Fprintf(w, "quickcmd_agent_jobs_failed %d\n", failed)
}

// executeJob executes a job in the background
func (s *Server) executeJob(ctx context.Context, job *Job) {
	job.Status = JobStatusRunning
	
	result, err := s.executor.Execute(ctx, job.Payload, job.LogChan)
	if err != nil {
		job.Status = JobStatusFailed
		result.Status = JobStatusFailed
		result.Error = err.Error()
	} else {
		job.Status = JobStatusCompleted
		result.Status = JobStatusCompleted
	}
	
	job.Result = result
	
	// Send final log frame
	job.LogChan <- &LogFrame{
		JobID:     job.Payload.JobID,
		Timestamp: time.Now(),
		Final:     true,
	}
	close(job.LogChan)
}

// Helper functions

func (s *Server) isAllowedController(controllerID string) bool {
	for _, allowed := range s.config.AllowedControllers {
		if allowed == controllerID {
			return true
		}
	}
	return false
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	response := map[string]interface{}{
		"error": message,
	}
	
	if err != nil {
		response["details"] = err.Error()
	}
	
	json.NewEncoder(w).Encode(response)
}
