package tunnel

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// MockProvider implements Provider interface for testing purposes.
type MockProvider struct {
	connectedCalled bool
	connectPort     int
	closeCalled     bool
}

// implement Provider interface
func (m *MockProvider) Connect(ctx context.Context, localPort int) (string, error) {
	m.connectedCalled = true
	m.connectPort = localPort
	return "https://abc123.example.com", nil
}

func (m *MockProvider) Close() error {
	m.closeCalled = true
	return nil
}

func (m *MockProvider) IsConnected() bool {
	return m.connectedCalled && !m.closeCalled
}

func (m *MockProvider) PublicURL() string {
	return "https://abc123.example.com"
}

func (m *MockProvider) Name() string {
	return "MockProvider"
}

// NewService creates a service with the MockProvider for testing.
func TestNewService(t *testing.T) {
	mockProvider := &MockProvider{}
	svc := NewService(mockProvider)

	if svc == nil {
		t.Fatal("Service should not be nil")
	}
}

// flakyProvider fails the first Connect call and succeeds afterwards.
type flakyProvider struct {
	MockProvider
	calls int
}

func (f *flakyProvider) Connect(ctx context.Context, localPort int) (string, error) {
	f.calls++
	if f.calls == 1 {
		return "", fmt.Errorf("temporary network error")
	}
	return f.MockProvider.Connect(ctx, localPort)
}

func TestService_StartRetryAfterFailure(t *testing.T) {
	svc := NewService(&flakyProvider{})
	ctx := context.Background()

	if err := svc.Start(ctx, 3000); err == nil {
		t.Fatal("first Start() error = nil, want connect error")
	}

	if err := svc.Start(ctx, 3000); err != nil {
		t.Fatalf("retry Start() error = %v, want nil", err)
	}
}

func TestService_Start(t *testing.T) {
	mock := &MockProvider{}

	svc := NewService(mock)
	ctx := context.Background()
	port := 3000
	err := svc.Start(ctx, port)
	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	// check mock.Connect was called
	if !mock.connectedCalled {
		t.Error("provider.Connect was not called")
	}

	// check port is correct
	if mock.connectPort != port {
		t.Errorf("connectPort = %d, want %d", port, mock.connectPort)
	}

	// Check Ready() channel is closed
	select {
	case <-svc.Ready():
	// Ok! channel is closed
	default:
		t.Error("ready channel should be closed after Start()")
	}

}

func TestService_PublicURL(t *testing.T) {
	mock := &MockProvider{}
	svc := NewService(mock)

	url := svc.PublicURL()
	if url != mock.PublicURL() {
		t.Errorf("Expected %s, got %s", mock.PublicURL(), svc.PublicURL())
	}
}

func TestService_Ready(t *testing.T) {
	mock := &MockProvider{}
	svc := NewService(mock)

	// Test before Start (Read channel not closed)
	select {
	case <-svc.Ready():
		t.Error("ready before even `Start` is called!")
	default:
		// alright
	}

	ctx := context.Background()

	// Test after Start (Ready channel closed)
	err := svc.Start(ctx, 3000)
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}

	select {
	case <-svc.Ready():
	// closed as expected
	default:
		// error if channel is not closed to send ready signal
		t.Error("ready channel should have been closed")
	}
}

func TestService_Close(t *testing.T) {
	mock := &MockProvider{}

	svc := NewService(mock)

	// before calling Close
	if mock.closeCalled {
		t.Error("provider should not be closed yet")
	}

	err := svc.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	if !mock.closeCalled {
		t.Error("provider.Close() was not called")
	}
}

func TestService_StartTwice(t *testing.T) {
	mock := &MockProvider{}
	svc := NewService(mock)

	ctx := context.Background()

	// First Start - should succeed
	err := svc.Start(ctx, 3000)

	if err != nil {
		t.Fatalf("First Start() error = %v, want nil", err)
	}

	err = svc.Start(ctx, 3000)
	if err == nil {
		t.Fatalf("Second start shall cause error")
	}

	if !strings.Contains(err.Error(), "already started") {
		t.Error("already started error shall be returned")
	}
}

func TestService_Providername(t *testing.T) {
	mock := &MockProvider{}
	svc := NewService(mock)

	if got := svc.ProviderName(); got != "MockProvider" {
		t.Errorf("ProviderName() = %s, want MockProvider", got)
	}
}

func TestService_IsConnected(t *testing.T) {
	mock := &MockProvider{}
	svc := NewService(mock)

	if svc.IsConnected() {
		t.Error("expected IsConnected to be false before Start()")
	}

	_ = svc.Start(context.Background(), 3000)
	if !svc.IsConnected() {
		t.Error("expected IsConnected to be true after Start()")
	}

	_ = svc.Close()
	if svc.IsConnected() {
		t.Error("expected IsConnected to be false after Close()")
	}
}

func TestService_WaitReady(t *testing.T) {
	mock := &MockProvider{}
	svc := NewService(mock)

	// Before start, WaitReady with short timeout should timeout
	err := svc.WaitReady(50 * time.Millisecond)
	if err == nil {
		t.Error("expected timeout error from WaitReady before start")
	}

	// After start, WaitReady returns nil immediately
	_ = svc.Start(context.Background(), 3000)
	err = svc.WaitReady(time.Second)
	if err != nil {
		t.Errorf("expected WaitReady to return nil after start, got: %v", err)
	}
}
