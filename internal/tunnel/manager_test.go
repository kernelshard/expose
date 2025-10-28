package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	port := 3000
	m := NewManager(port)

	if m == nil {
		t.Fatal("Expected Manager instance, got nil")
	}

	if m.localPort != port {
		t.Fatalf("Expected localPort to be %d, got %d", port, m.localPort)
	}

	if m.ready == nil {
		t.Errorf("ready channel not initialized")
	}

	// PublicURL should be empty initially
	if m.PublicURL() != "" {
		t.Errorf("Expected empty PublicURL, got %s", m.PublicURL())
	}
}

// TestManager_PublicURL verifies thread-safe access to PublicURL.
func TestManager_PublicURL(t *testing.T) {
	m := NewManager(3000)

	// initially empty
	if url := m.PublicURL(); url != "" {
		t.Errorf("expected empty URL, got %s", url)
	}

	// Check URL (simulate Start behavior)
	m.mu.Lock()
	m.publicURL = "http://localhost:8080"
	m.mu.Unlock()

	if url := m.PublicURL(); url != "http://localhost:8080" {
		t.Errorf("expected URL to be http://localhost:8080, got %s", url)
	}
}

// TestManager_Ready verifies the Ready channel behavior.
func TestManager_Ready(t *testing.T) {
	m := NewManager(3000)

	// Initially should block
	select {
	case <-m.Ready():
		t.Error("read channel closed before Start()")
	default:
		// expected: channel is open
	}

	// Close it (simulate Start behavior)
	close(m.ready)

	// Now it should be closed & non-blocking
	select {
	case <-m.Ready():
		// expected: channel is closed
	case <-time.After(110 * time.Millisecond):
		t.Error("read channel did not close after Start()")
	}

}

// TestManager_Start_Success tests successful Start of the Manager.
func TestManager_Start_Success(t *testing.T) {
	m := NewManager(3000)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	// Start in goroutine to avoid blocking
	go func() {
		errCh <- m.Start(ctx)
	}()

	// Wait for Ready
	select {
	case <-m.Ready():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for Ready signal from Start()")
	}

	// Verify publicURL is set
	publicURL := m.PublicURL()
	if publicURL == "" {
		t.Error("expected PublicURL to be set after Start()")
	}

	t.Logf("Tunnel URL: %s", publicURL)

	// cancel & wait for graceful shutdown
	cancel()

	select {
	case err := <-errCh:
		// Expect no error on normal shutdown
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("expected nil error on shutdown, got %v", err)
		}
	case <-time.After(700 * time.Millisecond):
		t.Error("timeout waiting for Start() to return after context cancellation")
	}
}

// TestManager_Start_PreCancelledContext veries early exit on cancelled context.
func TestManager_Start_PreCancelledContext(t *testing.T) {
	m := NewManager(3000)
	// Pre-cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Start should return immediately with context.Canceled error
	err := m.Start(ctx)
	if err == nil {
		t.Error("expected error from pre-cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

// TestManager_Close verifies resource cleanup on Close.
func TestManager_Close(t *testing.T) {
	m := NewManager(3000)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the manager
	go m.Start(ctx)

	// Wait for Ready
	<-m.Ready()

	// Close should not return on first call
	if err := m.Close(); err != nil {
		t.Errorf("expected nil error on Close(), got %v", err)
	}

	// Subsequent close should be safe
	_ = m.Close()

}

// TestManager_Close_BeforeStart verifies close is safe before Start is called.
func TestManager_Close_BeforeStart(t *testing.T) {
	m := NewManager(3000)

	// Close before start should not panic or error
	if err := m.Close(); err != nil {
		t.Errorf("expected nil error on Close() before Start(), got %v", err)
	}
}

// TestManager_InterfaceCompliance
func TestManager_InterfaceCompliance(t *testing.T) {
	var _ Tunneler = (*Manager)(nil)
}

// TestManager_ConcurrentURLAccess verifies thread safety
func TestManager_ConcurrentURLAccess(t *testing.T) {
	m := NewManager(3000)

	// Concurrent writest
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.mu.Lock()
			m.publicURL = fmt.Sprintf("http://localhost:%d", 8000+i)
			m.mu.Unlock()
		}(i)
	}

	// Concurrent reads
	for range 10 {
		wg.Go(func() {
			_ = m.PublicURL()
		})
	}

	wg.Wait()
}

// TestManager_ProxyHandler_NoLocalServer verifies error handling when local server is down.
func TestManager_ProxyHandler_NoLocalServer(t *testing.T) {
	m := NewManager(65000) // assuming nothing is running on this high port

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	m.proxyHandler(w, req)

	// It should return 502 Bad Gateway if local server is unreachable
	if w.Code != http.StatusBadGateway {
		t.Errorf("expected status 502 Bad Gateway, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected error message in response body, got empty body")
	}
	t.Logf("Error response: %s", body)
}

// TestManager_ProxyHandler_WithLocalServer tests the proxy handler when the local server is running.
func TestManager_ProxyHandler_WithLocalServer(t *testing.T) {
	// values for verification
	testBodyStr := "Hello from local server"
	testHeaderStr := "X-Test-Header"
	testHeaderValue := "test-value"

	// Create a test local server
	localServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(testHeaderStr, testHeaderValue)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testBodyStr))
	}))

	// closing is must to avoid resource leaks
	defer localServer.Close()

	// Extract port from test server
	_, portStr, _ := net.SplitHostPort(localServer.Listener.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	// Create manager pointing to the test local server port to simulate proxying
	m := NewManager(port)

	// Test the proxy handler
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	m.proxyHandler(w, req)

	// It should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w.Code)
	}

	if w.Header().Get(testHeaderStr) != testHeaderValue {
		t.Errorf(
			"expected %s to be '%s', got '%s'", testHeaderStr, testHeaderValue,
			w.Header().Get(testHeaderStr),
		)
	}

	body := w.Body.String()
	if body != testBodyStr {
		t.Errorf("expected body '%s', got '%s'", testBodyStr, body)
	}

}

func TestManager_FullLifeCycle(t *testing.T) {
	m := NewManager(3000)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the manager
	go m.Start(ctx)

	// Wait for Ready

	select {
	case <-m.Ready():
		// expected
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for Ready signal from Start()")
	}

	// Verify URL is set
	if m.PublicURL() == "" {
		t.Error("publicURL not set after Start()")
	}

	// // Close explicitly
	// if err := m.Close(); err != nil {
	// 	t.Errorf("error on Close(): %v", err)
	// }
	err := m.Close()

	// Note: Close() may return "use of closed network connection"
	// because server.Close() already closed the listener.
	// This is expected behavior in errors.Join()
	if err != nil {
		t.Errorf("unexpected error on Close(): %v", err)
	}
}
