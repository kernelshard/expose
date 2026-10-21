package pool

import (
	"bytes"
	"io"
	"net"
	"testing"
)

type devNull struct{}

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkCopy_Unpooled(b *testing.B) {
	b.ReportAllocs()

	// 1. Allocate test data ONCE outside the loop so we don't benchmark make()
	data := make([]byte, 100*1024) // 100 KB

	for b.Loop() {
		client, server := net.Pipe()

		// 2. Stream data in a background goroutine and close to signal EOF
		go func() {
			client.Write(data)
			client.Close()
		}()

		// 3. io.Copy reads until EOF
		_, err := io.Copy(devNull{}, server)
		if err != nil {
			b.Fatalf("io.Copy error: %v", err)
		}
		server.Close()
	}
}

func BenchmarkCopy_Pooled_WithBuffer(b *testing.B) {
	b.ReportAllocs()

	data := make([]byte, 100*1024) // 100 KB

	for b.Loop() {
		client, server := net.Pipe()

		go func() {
			client.Write(data)
			client.Close()
		}()

		_, err := Copy(devNull{}, server)
		if err != nil {
			b.Fatalf("pool.Copy error: %v", err)
		}
		server.Close()
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"Empty", 0},
		{"Small", 1024},
		{"LargeThanBuffer", 128 * 1024},
	}
	for _, tc := range tests {
		data := bytes.Repeat([]byte("a"), tc.size)
		src := bytes.NewReader(data)

		var dst bytes.Buffer

		n, err := Copy(&dst, src)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}
		if n != int64(tc.size) {
			t.Errorf("%s: expected %d bytes, got %d", tc.name, tc.size, n)
		}
		if !bytes.Equal(dst.Bytes(), data) {
			t.Errorf("%s: copied data does not match original", tc.name)
		}

	}
}
