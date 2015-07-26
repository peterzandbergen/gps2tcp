package gps2tcp

import (
	"io"
	"log"
)

type chanToWriterCloser struct {
	id        int
	in        <-chan []byte
	out       io.WriteCloser
	closeChan chan<- int
}

// LoopTcp writes the bytes it receives on bytesChan to wc.
func LoopTcp(cw *chanToWriterCloser) {
	for b := range cw.in {
		// log.Printf("LoopTcp: Bytes arrived: %d.\n", len(b))
		n, err := cw.out.Write(b)
		_ = n
		// log.Printf("LoopTcp: Bytes written: %d.\n", n)
		putBuffer(b)
		if err != nil {
			// log.Printf("LoopTcp: Error writing bytes: %s.\n", err.Error())
			cw.closeChan <- cw.id
		}
	}
	log.Printf("LoopTcp: Exiting.\n")
}
