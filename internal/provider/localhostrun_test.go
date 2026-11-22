package provider

import (
	"crypto/rand"
	"errors"
	"testing"

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
