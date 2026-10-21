// Package pool provides reusable buffers for streaming.
package pool

import (
	"io"
	"sync"
)

// BufferSize is 32 KB, matching standard io.Copy and L1 cache size
const BufferSize = 32 * 1024

var copyBufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, BufferSize)
		return &b
	},
}

// Copy copies from src to dst using a reused 32 KB buffer
func Copy(dst io.Writer, src io.Reader) (int64, error) {
	bufPtr := copyBufferPool.Get().(*[]byte)

	defer copyBufferPool.Put(bufPtr)

	return io.CopyBuffer(dst, src, *bufPtr)
}
