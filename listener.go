package gps2tcp

import (
	"io"
	"log"
	"net"
)

type listenerToChan struct {
	addr   string
	done   <-chan struct{}
	wcChan chan<- io.WriteCloser
}

func (lc *listenerToChan) Run() {
	// Start the Listener.
	// Accept connections.
	// Send connections to wcChan

	l, err := net.Listen("tcp", lc.addr)
	if err != nil {
		close(lc.wcChan)
		return
	}
	defer l.Close()

	for {
		// Wait for a new connection.
		c, err := l.Accept()
		log.Printf("LoopListener: Accept.\n")
		if err != nil {
			log.Printf("LoopListener: Accept error: %s.\n", err.Error())
			close(lc.wcChan)
			return
		}

		select {
		case lc.wcChan <- c:
			// Send the connection to the central server.
			c = nil

		case <-lc.done:
			// Stops the downstream process.
			close(lc.wcChan)
			log.Println("LoopListener: done closed.")
			return
		}
	}
}

// LoopListener opens a listener on the addr and sends connections that arrive
// to the wcChan as WriterClosers.
//func LoopListener(done <-chan struct{}, addr string, wcChan chan<- io.WriteCloser) {
//	// Start the Listener.
//	// Accept connections.
//	// Send connections to wcChan

//	l, err := net.Listen("tcp", addr)
//	if err != nil {
//		close(wcChan)
//		return
//	}
//	defer l.Close()

//	for {
//		// Wait for a new connection.
//		c, err := l.Accept()
//		log.Printf("LoopListener: Accept.\n")
//		if err != nil {
//			log.Printf("LoopListener: Accept error: %s.\n", err.Error())
//			close(wcChan)
//			return
//		}

//		select {
//		case wcChan <- c:
//			// Send the connection to the central server.
//			c = nil

//		case <-done:
//			// Stops the downstream process.
//			close(wcChan)
//			log.Println("LoopListener: done closed.")
//			return
//		}
//	}
//}
