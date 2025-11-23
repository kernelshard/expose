package provider

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sync"
	"time"
)

// Cloudflare implements the Provider interface for Cloudflare Tunnel
type Cloudflare struct {
	cmd       *exec.Cmd
	mu        sync.RWMutex
	publicURL string

	// RequestTunnel is exported for test mocking
	RequestTunnel func(ctx context.Context, port int, timeout time.Duration) (string, *exec.Cmd, error)
}

// NewCloudFlare creates a new instance of Cloudflare provider
func NewCloudFlare() *Cloudflare {
	return &Cloudflare{
		RequestTunnel: requestTunnel, // Use real implementation by default
	}
}

// NewCloudFlareWithMock creates a Cloudflare provider with a mock requestTunnel function for testing
func NewCloudFlareWithMock(mockRequestTunnel func(ctx context.Context, port int, timeout time.Duration) (string, *exec.Cmd, error)) *Cloudflare {
	return &Cloudflare{
		RequestTunnel: mockRequestTunnel,
	}
}

// Connect establishes a Cloudflare Tunnel to the specified local port
func (c *Cloudflare) Connect(ctx context.Context, localPort int) (string, error) {
	timeout := 30 * time.Second
	url, cmd, err := c.RequestTunnel(ctx, localPort, timeout)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.cmd = cmd
	c.publicURL = url
	c.mu.Unlock()

	return url, nil
}

// Close terminates the Cloudflare Tunnel
func (c *Cloudflare) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// if cmd is running, kill the process
	if c.cmd != nil && c.cmd.Process != nil {
		err := c.cmd.Process.Kill()
		// clear fields safely under write lock
		c.cmd = nil
		c.publicURL = ""
		return err
	}
	return nil
}

// PublicURL returns the public URL of the Cloudflare Tunnel
func (c *Cloudflare) PublicURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.publicURL
}

// Name returns the name of the provider
func (c *Cloudflare) Name() string {
	return "Cloudflare"
}

// requestTunnel starts the cloudflared process and retrieves the public URL
func requestTunnel(ctx context.Context, port int, timeout time.Duration) (string, *exec.Cmd, error) {
	urlRegex := regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", fmt.Sprintf("http://localhost:%d", port))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", nil, fmt.Errorf("get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("start cloudflared: %w", err)
	}

	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Read stderr for URL
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line) // logs

			if url := urlRegex.FindString(line); url != "" {
				urlCh <- url
				return
			}
		}

		// Handle scanner error or no URL found
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("read stderr: %w", err)
		} else {
			errCh <- fmt.Errorf("cloudflared exited without providing URL")
		}
	}()

	// Wait for result with timeout
	select {
	case url := <-urlCh:
		// Success - return cmd so caller can manage it
		return url, cmd, nil

	case err := <-errCh:
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", nil, err

	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", nil, fmt.Errorf("timeout waiting for tunnel URL")

	case <-ctx.Done():
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return "", nil, ctx.Err()
	}
}
