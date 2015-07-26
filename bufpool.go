package gps2tcp

import (
	_ "log"
)

const (
	bufferSize     = 100
	bufferPoolSize = 100
)

// Buffer pool to attempt reuse of the buffers.

var bufferPool = make(chan []byte, bufferPoolSize)
var nbuf int
var created int
var dropped int

// GetBuffer tries to get a buffer from the pool.
// It creates a new buffer is non is available.
func getBuffer(msg ...string) []byte {

	var b []byte
	select {
	case b = <-bufferPool:
		// Size b up.
		nbuf--
		// log.Printf("BufferPool: reused existing buffer: %d.\n", nbuf)
		b = b[0:cap(b)]

	default:
		// log.Printf("BufferPool: created new buffer: %d.\n", nbuf)
		created++
		b = make([]byte, bufferSize)
	}

	// log.Printf("getBuffer (%d, %d, %d): %s\n", created, dropped, nbuf, msg)
	return b
}

func putBuffer(b []byte, msg ...string) {

	select {
	case bufferPool <- b:
		nbuf++
		// log.Printf("BufferPool: returned buffer to pool: %d.\n", nbuf)

	default:
		dropped++
		// log.Printf("BufferPool: dropped buffer: %d.\n", nbuf)
	}
	// log.Printf("putBuffer (%d, %d, %d): %s\n", created, dropped, nbuf, msg)
}
