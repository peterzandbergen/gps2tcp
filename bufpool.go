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

// GetBuffer tries to get a buffer from the pool.
// It creates a new buffer is none is available.
func getBuffer(msg ...string) []byte {

	var b []byte
	select {
	case b = <-bufferPool:
		// Size b up.
		b = b[0:cap(b)]

	default:
		b = make([]byte, bufferSize)
	}

	return b
}

func putBuffer(b []byte, msg ...string) {

	select {
	case bufferPool <- b:

	default:
	}
}
