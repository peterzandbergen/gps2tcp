package gps2tcp

import (
	"github.com/tarm/goserial"
	"io"
	"log"
	"os"
	"time"
)

const (
	TimeOut    = 2 * time.Second
	Retries    = 1
	RetrySleep = 2 * time.Second
	BaudRate   = 4800
)

// serialToChan
type serialToChan struct {
	// The serial port, COM5 for me.
	port string
	// Channel that Run sends the read bytes to.
	bytesChan chan<- []byte
	// done is closed to indicate that Run can stop.
	done <-chan struct{}
}

// Run starts reading bytes from the serial port, (re)opening it if needed
// and sends the read bytes to the channel.
// Run blocks so should be called as a go routine.
func (sc *serialToChan) Run() {
	// Ser is nil when closed.
	var ser io.ReadWriteCloser

	for {
		// Open the serial port.
		if ser == nil {
			// Open serial port.
			log.Printf("LoopSerial: Opening the port.\n")
			var err error
			// Open the serial port.
			ser, err = openSerial(sc.port)
			if err != nil {
				log.Printf("LoopSerial: Error opening the port: %s.\n", err.Error())
				ser = nil
				time.Sleep(5 * time.Second)
			}
		}

		// b is nil if no bytes were received.
		var b []byte

		if ser != nil {
			// Read buf from serial.
			b = getBuffer("LoopSerial: asking for buffer.")
			n, err := ser.Read(b)
			if err != nil || n == 0 {
				// Error reading the serial port.
				log.Printf("LoopSerial: Error reading the serial port. Closing it.")
				// Return and invalidate the buffer.
				putBuffer(b, "LoopSerial: returning buffer because of read error.")
				b = nil
				// Close serial.
				ser.Close()
				ser = nil
			} else {
				b = b[:n]
			}
		}

		if b != nil {
			// Send the bytes to the channel.
			sc.bytesChan <- b
			// Set b to nil to indicate that it has been sent.
			b = nil
		}

		// Check if the done channel has been closed, but don't wait.
		select {
		case <-sc.done:
			// Time to stop.
			if ser != nil {
				ser.Close()
			}
			close(sc.bytesChan)
			return
		default:
		}
	}
}

// OpenSerial opens the serial port with name and returns an interface or an error.
// It tries Retries times to open the port to add some time for the port to be ready.
func openSerial(name string) (io.ReadWriteCloser, error) {
	var err error
	// First check 5 if the port exists.
	log.Printf("Checking for port presence.\n")
	_, err = os.Stat(name)
	if err != nil {
		return nil, err
	}

	// Open the port.
	log.Printf("Opening the port.\n")
	cfg := &serial.Config{
		Name:        name,
		Baud:        BaudRate,
		ReadTimeout: TimeOut,
	}
	f, err := serial.OpenPort(cfg)
	if err != nil {
		log.Printf("Failed to open port %s: %s\n", name, err.Error())
		return nil, err
	}
	log.Printf("Port opened.\n")
	return f, nil
}
