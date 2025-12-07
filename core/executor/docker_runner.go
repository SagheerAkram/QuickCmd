package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// SandboxOptions configures sandbox execution
type SandboxOptions struct {
	WorkingDir    string        // Working directory inside container
	Mounts        []Mount       // Volume mounts
	NetworkAccess bool          // Enable network access
	CPULimit      float64       // CPU limit (cores, e.g., 0.5)
	MemoryLimit   int64         // Memory limit in bytes
	PidsLimit     int64         // Max number of processes
	Timeout       time.Duration // Execution timeout
	Image         string        // Docker image to use
	ReadOnly      bool          // Mount filesystem as read-only
}

// Mount represents a volume mount
type Mount struct {
	Source   string
	Target   string
	ReadOnly bool
}

// SandboxResult contains execution results
type SandboxResult struct {
	Stdout     []byte
	Stderr     []byte
	ExitCode   int
	SandboxID  string
	StartTime  time.Time
	EndTime    time.Time
	Error      error
}

// DockerRunner executes commands in Docker containers
type DockerRunner struct {
	client *client.Client
}

// NewDockerRunner creates a new Docker runner
func NewDockerRunner() (*DockerRunner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w (is Docker installed and running?)", err)
	}
	
	// Verify Docker is accessible
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Docker daemon not accessible: %w", err)
	}
	
	return &DockerRunner{client: cli}, nil
}

// RunInSandbox executes a command in an isolated Docker container
func (dr *DockerRunner) RunInSandbox(cmd string, opts SandboxOptions) (*SandboxResult, error) {
	result := &SandboxResult{
		StartTime: time.Now(),
	}
	
	// Set defaults
	if opts.Image == "" {
		opts.Image = "alpine:latest"
	}
	if opts.CPULimit == 0 {
		opts.CPULimit = 0.5
	}
	if opts.MemoryLimit == 0 {
		opts.MemoryLimit = 256 * 1024 * 1024 // 256MB
	}
	if opts.PidsLimit == 0 {
		opts.PidsLimit = 64
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Minute
	}
	if opts.WorkingDir == "" {
		opts.WorkingDir = "/workspace"
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	
	// Ensure image exists
	if err := dr.ensureImage(ctx, opts.Image); err != nil {
		result.Error = err
		return result, err
	}
	
	// Create container config
	containerConfig := &container.Config{
		Image:      opts.Image,
		Cmd:        []string{"/bin/sh", "-c", cmd},
		WorkingDir: opts.WorkingDir,
		User:       "1000:1000", // Run as non-root
		Tty:        false,
		AttachStdout: true,
		AttachStderr: true,
	}
	
	// Host config with resource limits
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			NanoCPUs: int64(opts.CPULimit * 1e9),
			Memory:   opts.MemoryLimit,
			PidsLimit: &opts.PidsLimit,
		},
		AutoRemove: true, // Cleanup after execution
		ReadonlyRootfs: opts.ReadOnly,
	}
	
	// Network isolation
	if !opts.NetworkAccess {
		hostConfig.NetworkMode = "none"
	}
	
	// Add mounts
	for _, mount := range opts.Mounts {
		hostConfig.Binds = append(hostConfig.Binds, 
			fmt.Sprintf("%s:%s:%s", mount.Source, mount.Target, mountMode(mount.ReadOnly)))
	}
	
	// Create container
	resp, err := dr.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		result.Error = fmt.Errorf("failed to create container: %w", err)
		return result, result.Error
	}
	
	result.SandboxID = resp.ID[:12] // Short ID for display
	
	// Start container
	if err := dr.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		result.Error = fmt.Errorf("failed to start container: %w", err)
		return result, result.Error
	}
	
	// Wait for container to finish
	statusCh, errCh := dr.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			result.Error = fmt.Errorf("container wait error: %w", err)
			return result, result.Error
		}
	case status := <-statusCh:
		result.ExitCode = int(status.StatusCode)
	case <-ctx.Done():
		// Timeout - kill container
		dr.client.ContainerKill(context.Background(), resp.ID, "SIGKILL")
		result.Error = fmt.Errorf("execution timeout after %v", opts.Timeout)
		result.ExitCode = 124 // Standard timeout exit code
		return result, result.Error
	}
	
	// Get logs
	logReader, err := dr.client.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		result.Error = fmt.Errorf("failed to get container logs: %w", err)
		return result, result.Error
	}
	defer logReader.Close()
	
	// Separate stdout and stderr
	var stdout, stderr strings.Builder
	if _, err := stdcopy.StdCopy(&stdout, &stderr, logReader); err != nil {
		result.Error = fmt.Errorf("failed to read container logs: %w", err)
		return result, result.Error
	}
	
	result.Stdout = []byte(stdout.String())
	result.Stderr = []byte(stderr.String())
	result.EndTime = time.Now()
	
	return result, nil
}

// ensureImage pulls the image if it doesn't exist
func (dr *DockerRunner) ensureImage(ctx context.Context, image string) error {
	// Check if image exists locally
	_, _, err := dr.client.ImageInspectWithRaw(ctx, image)
	if err == nil {
		return nil // Image exists
	}
	
	// Pull image
	reader, err := dr.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	defer reader.Close()
	
	// Wait for pull to complete
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to download image %s: %w", image, err)
	}
	
	return nil
}

// mountMode returns the mount mode string
func mountMode(readOnly bool) string {
	if readOnly {
		return "ro"
	}
	return "rw"
}

// Close closes the Docker client
func (dr *DockerRunner) Close() error {
	if dr.client != nil {
		return dr.client.Close()
	}
	return nil
}

// IsDockerAvailable checks if Docker is available
func IsDockerAvailable() bool {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	defer cli.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	_, err = cli.Ping(ctx)
	return err == nil
}

// GetDockerInfo returns Docker system information
func GetDockerInfo() (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	info, err := cli.Info(ctx)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("Docker %s (%s)", info.ServerVersion, info.OperatingSystem), nil
}
