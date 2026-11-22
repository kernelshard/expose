package provider

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

type failingReader struct{}

func (f *failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestGenerateSSHKey(t *testing.T) {
	t.Run("success with rand.Reader", func(t *testing.T) {
		signer, err := generateSSHKey(rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if signer == nil {
			t.Fatal("expected non-nil signer")
		}
		if signer.PublicKey() == nil {
			t.Fatal("expected non-nil public key")
		}
		if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
			t.Fatalf("expected ed25519 key, got %s", signer.PublicKey().Type())
		}
	})

	t.Run("failure with failing reader", func(t *testing.T) {
		signer, err := generateSSHKey(&failingReader{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if signer != nil {
			t.Fatal("expected nil signer on error")
		}
	})
}

func TestLocalhostRun_GettersAndClose(t *testing.T) {
	p := NewLocalhostRun()

	if p.Name() != "localhostrun" {
		t.Fatalf("expected name localhostrun, got %s", p.Name())
	}

	if p.IsConnected() {
		t.Fatal("expected not connected initially")
	}

	if p.PublicURL() != "" {
		t.Fatalf("expected empty public URL, got %s", p.PublicURL())
	}

	// Close on uninitialized provider should be safe and idempotent
	if err := p.Close(); err != nil {
		t.Fatalf("unexpected error on close: %v", err)
	}

	if p.IsConnected() {
		t.Fatal("expected not connected after close")
	}
}

func TestLHRURLRegex(t *testing.T) {
	validCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "15:16:52 e8d1833a1abea5.app.lhr.life tunneled with tls termination, https://e8d1833a1abea5.app.lhr.life",
			expected: "https://e8d1833a1abea5.app.lhr.life",
		},
		{
			input:    "Your domain is: https://my-subdomain.lhr.life",
			expected: "https://my-subdomain.lhr.life",
		},
		{
			input:    "welcome http://test.lhr.pro active",
			expected: "http://test.lhr.pro",
		},
	}

	for _, tc := range validCases {
		match := lhrURLRegex.FindString(tc.input)
		if match != tc.expected {
			t.Errorf("input %q: expected match %q, got %q", tc.input, tc.expected, match)
		}
	}

	invalidCases := []string{
		"welcome to localhost.run",
		"https://example.com/not-lhr",
		"http://localhost:3000",
	}

	for _, input := range invalidCases {
		if match := lhrURLRegex.FindString(input); match != "" {
			t.Errorf("input %q: expected no match, got %q", input, match)
		}
	}
}

func TestLiveConnect(t *testing.T) {
	if os.Getenv("EXPOSE_LIVE_TESTS") == "" {
		t.Skip("skipping live network test; set EXPOSE_LIVE_TESTS=1 to run")
	}

	// 1. Start a local HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("HELLO FROM LOCAL EXPOSE"))
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen locally: %v", err)
	}
	defer listener.Close()

	server := &http.Server{Handler: mux}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Close()

	localPort := listener.Addr().(*net.TCPAddr).Port

	// 2. Connect localhostrun provider
	p := NewLocalhostRun()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	t.Log(time.Now().Format("15:04:05"), "Calling Connect...")
	url, err := p.Connect(ctx, localPort)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	t.Log(time.Now().Format("15:04:05"), "SUCCESS! URL:", url)
	defer p.Close()

	// 3. Make HTTP request through the public URL
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		t.Fatalf("GET to %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body failed: %v", err)
	}
	t.Log(time.Now().Format("15:04:05"), "Response Body:", string(body))
	if string(body) != "HELLO FROM LOCAL EXPOSE" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}

