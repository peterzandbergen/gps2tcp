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

type serialToChan struct {
	port      string
	bytesChan chan<- []byte
	done      <-chan struct{}
}

// Run blocks.
func (sc *serialToChan) Run() {
	// Ser is nil when closed.
	var ser io.ReadWriteCloser

	for {
		if ser == nil {
			// Open serial port.
			log.Printf("LoopSerial: Opening the port.\n")
			var err error
			// Open the serial port.
			ser, err = OpenSerial(sc.port)
			if err != nil {
				log.Printf("LoopSerial: Error opening the port: %s.\n", err.Error())
				ser = nil
			}
		}

		var outChan chan<- []byte

		if ser != nil {
			// Read buf from serial.
			b := getBuffer("LoopSerial: asking for buffer.")
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
				// Set outChan to nil because we have nothing to send.
				outChan = nil
			} else {
				outChan = sc.bytesChan
			}

			// Send the character to the channel, but don't wait.
			select {
			case outChan <- b[:n]:
				// nothing to do.
				// log.Printf("LoopSerial: Read %d bytes: %s\n", n, string(b))
				// log.Printf("LoopSerial: Read %d bytes.\n", n)
				b = nil

			case <-sc.done:
				// Time to stop.
				ser.Close()
				ser = nil
				close(sc.bytesChan)

				// default:
			}

		}
	}
}

// OpenSerial opens the serial port with name and returns an interface or an error.
// It tries 5 times to open the port to add some time for the port to be ready.
func OpenSerial(name string) (io.ReadWriteCloser, error) {
	var err error
	for i := 0; i < Retries; i++ {
		// First check 5 times if the port exists.
		log.Printf("Checking for port presence attempt %d.\n", i+1)
		_, err = os.Stat(name)
		if err != nil {
			time.Sleep(RetrySleep)
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	// Try to open the port.
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
