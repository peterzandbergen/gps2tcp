package gps2tcp

import (
	"io"
	"log"
	//"net"
	"sync"
)

type Server struct {
	// comport name
	comPort string
	// listen address
	address string

	// Channel where the bytes arrive.
	chanIn chan []byte
	// Channel that receives a WriterCloser to a socket.
	newChan chan io.WriteCloser
	// Channels that need to receive the bytes.
	chansOut map[int]chan []byte
	// Channel that receives the channels to be closed.
	closeChan chan int
	// Next channel id.
	chanID int
	// Done channel.
	done chan struct{}
}

// Prepare the server.
func NewServer(addr string, comPort string) *Server {
	var s = &Server{
		comPort: comPort,
		address: addr,
	}
	// Create the bytes in channel.
	s.chanIn = make(chan []byte, 10)
	// Fan out channels.
	s.chansOut = make(map[int]chan []byte, 10)
	// Close channel.
	s.closeChan = make(chan int, 10)
	// Done channel.
	s.done = make(chan struct{})
	// Channel for new connections.
	s.newChan = make(chan io.WriteCloser, 10)

	return s
}

func (s *Server) Stop() {
	close(s.done)
}

// Run services the incoming bytes and sends them to the outgoing channels.
// Outgoing channels can be added and removed.
// Run stops when the incoming channel is closed.
func (s *Server) Run(wg *sync.WaitGroup) {
	// Start the serial reader.
	sc := &serialToChan{
		done:      s.done,
		port:      s.comPort,
		bytesChan: s.chanIn,
	}
	go sc.Run()
	// Start the Listener.
	lc := &listenerToChan{
		addr:   s.address,
		done:   s.done,
		wcChan: s.newChan,
	}
	go lc.Run()

ForLoop:
	for {
		select {
		case wc := <-s.newChan:
			// Start a new LoopTcp.
			log.Printf("Run: New tcp channel arrived: %d.\n", s.chanID)
			nc := make(chan []byte)
			// Create new chan to writer.
			cw := &chanToWriterCloser{
				id:        s.chanID,
				in:        nc,
				out:       wc,
				closeChan: s.closeChan,
			}
			go cw.Run()
			log.Printf("Run: LoopTcp started: %d.\n", s.chanID)
			// Add connection to the map.
			s.chansOut[cw.id] = nc
			s.chanID++

		case cid := <-s.closeChan:
			// Fan out channel closed.
			log.Printf("Run: LoopTcp closing LoopTcp: %d.\n", cid)
			close(s.chansOut[cid])
			delete(s.chansOut, cid)
			log.Printf("Run: number of consumers left: %d.\n", len(s.chansOut))

		case b, ok := <-s.chanIn:
			// Buffer with bytes received, or the channel closed.
			// log.Printf("Run: Bytes arrived: %d.\n", len(b))
			if !ok {
				// Close all outgoing channels and stop.
				log.Println("Chan in closed.")
				for _, c := range s.chansOut {
					close(c)
				}
				break ForLoop
			}
			// Fan out the buffer.
			for _, c := range s.chansOut {
				// Get a new buffer, copy the bytes and send on.
				nb := getBuffer("Run: asking for buffer for fan out.")

				nc := copy(nb, b)
				c <- nb[:nc]
			}
			// Return the buffer to the pool.
			putBuffer(b, "Run: returning buffer from serial.")
		}
	}
	if wg != nil {
		wg.Done()
	}
}
