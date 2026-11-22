package provider

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

const defaultLocalhostRunServer = "localhost.run:22"

func generateSSHKey(random io.Reader) (ssh.Signer, error) {
	_, privKey, err := ed25519.GenerateKey(random)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(privKey)
}

type LocalhostRun struct {
	serverAddr string
	client     *ssh.Client
	publicURL  string
	connected  bool
	mu         sync.RWMutex
}

// NewLocalhostRun creates a new localhost.run provider
func NewLocalhostRun() *LocalhostRun {
	return &LocalhostRun{
		serverAddr: defaultLocalhostRunServer,
	}
}

func (l *LocalhostRun) Name() string {
	return "localhostrun"
}

func (l *LocalhostRun) IsConnected() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.connected
}

func (l *LocalhostRun) PublicURL() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.publicURL
}

// dialSSH establishes a SSH connection with full context cancellation support.
// Standard ssh.Dial ignores context.Context, causing CLI freezes when a user presses Ctrl+C
// This helper cancles both the initial TCP dial and any hung SSH cryptographic handshake
func dialSSH(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	// 1. Dial the raw TCP socket, aborting immediately if ctx is canceled
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	// 2. If ctx is canceled while the SSH handshake is waiting for the server
	// close the raw socket so NewClientConn unblocks instead of freezing
	stop := context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})
	defer stop()

	// 3. Negotiate encryption keys and aytgebtucate over the open socket
	sshConn, incomingStreams, serverRequests, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 4. client that wraps the entire SSH session
	client := ssh.NewClient(sshConn, incomingStreams, serverRequests)

	return client, nil
}

// Connect establishes an SSH connection to localhost.run with reverse port forwarding.
func (l *LocalhostRun) Connect(ctx context.Context, localPort int) (string, error) {
	// 1. Generate ephemeral keycard
	signer, err := generateSSHKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate SSH key: %w", err)
	}

	// 2. Configure SSH client
	config := &ssh.ClientConfig{
		User: "nokey",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	// 3. Dial TCP respecting context cancellation
	client, err := dialSSH(ctx, "tcp", l.serverAddr, config)
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %w", l.serverAddr, err)
	}

	l.mu.Lock()
	l.client = client
	l.connected = true
	l.mu.Unlock()

	return "", nil // Temporary: Phase 3 will return the public URL!}
}

func (l *LocalhostRun) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.connected = false
	l.publicURL = ""

	if l.client != nil {
		err := l.client.Close()
		l.client = nil
		return err
	}

	return nil
}
