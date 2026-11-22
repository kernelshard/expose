package provider

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/kernelshard/expose/internal/pool"
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
	channel    ssh.Channel
	listener   net.Listener
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

	// 4. Request reverse port forwarding on remote port 80
	listener, err := client.Listen("tcp", ":80")
	if err != nil {
		_ = client.Close()
		return "", fmt.Errorf("failed to request reverse port: %w", err)
	}

	// 5. Read assigned public URL from the SSH banner
	publicURL, channel, err := extractPublicURL(ctx, client)
	if err != nil {
		_ = listener.Close()
		_ = client.Close()
		return "", err
	}

	l.mu.Lock()
	l.client = client
	l.channel = channel
	l.listener = listener
	l.publicURL = publicURL
	l.connected = true
	l.mu.Unlock()

	// 6. Start forwarding visitor traffic to the local port
	go l.forwardTraffic(ctx, listener, localPort)

	return publicURL, nil
}

// forwardTraffic continuously accepts incoming visitor connections from the SSH listener.
func (l *LocalhostRun) forwardTraffic(ctx context.Context, listener net.Listener, localPort int) {
	for {
		remoteConn, err := listener.Accept()
		if err != nil {
			return
		}
		go l.proxyConnection(ctx, remoteConn, localPort)
	}
}

// proxyConnection bridges a single incoming visitor connection to the local port.
func (l *LocalhostRun) proxyConnection(ctx context.Context, remoteConn net.Conn, localPort int) {
	defer remoteConn.Close()

	var dialer net.Dialer
	localConn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return
	}
	defer localConn.Close()

	done := make(chan struct{}, 2)
	go func() {
		_, _ = pool.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()
	go func() {
		_, _ = pool.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()
	<-done
}

func (l *LocalhostRun) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.connected = false
	l.publicURL = ""

	if l.channel != nil {
		_ = l.channel.Close()
		l.channel = nil
	}

	if l.listener != nil {
		_ = l.listener.Close()
		l.listener = nil
	}

	if l.client != nil {
		err := l.client.Close()
		l.client = nil
		return err
	}

	return nil
}

var lhrURLRegex = regexp.MustCompile(`https?://[a-zA-Z0-9-.]+\.(?:lhr\.life|lhr\.pro|lhrtunnel\.link)`)

const maxBannerBytes = 32 * 1024

// extractPublicURL opens a session channel, requests a shell without waiting for a reply
// (localhost.run's russh server does not send channel request confirmations),
// and reads the banner until it finds our https://...lhr.life URL.
func extractPublicURL(ctx context.Context, client *ssh.Client) (string, ssh.Channel, error) {
	ch, inReqs, err := client.OpenChannel("session", nil)
	if err != nil {
		return "", nil, fmt.Errorf("open ssh channel: %w", err)
	}
	go ssh.DiscardRequests(inReqs)

	stop := context.AfterFunc(ctx, func() {
		_ = ch.Close()
	})
	defer stop()

	// wantReply=false because localhost.run server does not acknowledge channel requests
	if _, err := ch.SendRequest("shell", false, nil); err != nil {
		_ = ch.Close()
		return "", nil, fmt.Errorf("request shell: %w", err)
	}

	buf := make([]byte, 1024)
	var output strings.Builder
	totalBytes := 0

	for {
		n, err := ch.Read(buf)
		if n > 0 {
			totalBytes += n
			if totalBytes > maxBannerBytes {
				_ = ch.Close()
				return "", nil, fmt.Errorf("exceeded max banner read limit (%d bytes)", maxBannerBytes)
			}
			output.WriteString(string(buf[:n]))
			if match := lhrURLRegex.FindString(output.String()); match != "" {
				return match, ch, nil
			}
		}
		if err != nil {
			_ = ch.Close()
			return "", nil, fmt.Errorf("reading banner: %w", err)
		}
	}
}
