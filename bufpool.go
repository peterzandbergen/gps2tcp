package gps2tcp

const (
	bufferSize     = 100
	bufferPoolSize = 100
)

// Buffer pool to attempt reuse of the buffers.

var bufferPool = make(chan []byte, bufferPoolSize)

// GetBuffer tries to get a buffer from the pool.
// It creates a new buffer is non is available.
func getBuffer() []byte {
	var b []byte
	select {
	case b = <-bufferPool:
		// Size b up.
		return b[0:cap(b)]
	default:
		return make([]byte, bufferSize)
	}
}

func putBuffer(b []byte) {
	select {
	case bufferPool <- b:
	default:
	}
}
