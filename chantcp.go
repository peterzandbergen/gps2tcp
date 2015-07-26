package gps2tcp

import (
	"io"
	"log"
)

type chanToWriterCloser struct {
	// Channel identifier.
	id int
	// Channel receiving the buffers.
	in  <-chan []byte
	out io.WriteCloser
	// Channel to request a close.
	closeChan chan<- int
}

// LoopTcp writes the bytes it receives on bytesChan to wc.
func (cw *chanToWriterCloser) Run() {
	for b := range cw.in {
		// log.Printf("LoopTcp: Bytes arrived: %d.\n", len(b))
		n, err := cw.out.Write(b)
		putBuffer(b, "LoopTcp: returning buffer after writing to socket.")
		_ = n
		// log.Printf("LoopTcp: Bytes written: %d.\n", n)
		if err != nil {
			// log.Printf("LoopTcp: Error writing bytes: %s.\n", err.Error())
			// Send a signal to get closed.
			cw.closeChan <- cw.id
			break
		}
	}
	// Keep reading till the channel is closed.
	for b := range cw.in {
		putBuffer(b, "LoopTcp: returning buffer after writing to socket.")
	}
	log.Printf("LoopTcp: Exiting.\n")
}
