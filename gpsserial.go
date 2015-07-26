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

func LoopSerialChannel(done <-chan struct{}, port string, bytesChan chan<- []byte) {

	var ser io.ReadWriteCloser
	for {
		// Open serial port.
		if ser == nil {
			log.Printf("LoopSerial: Opening the port.\n")
			var err error
			// Open the serial port.
			ser, err = OpenSerial(port)
			if err != nil {
				log.Printf("LoopSerial: Error opening the port: %s.\n", err.Error())
				ser = nil
			}
			select {
			case <-done:
				log.Println("Done is closed.")
				close(bytesChan)
				return
			default:
			}
		} else {
			// log.Printf("LoopSerial: Starting to read.\n")
			// Read buf from serial.
			b := getBuffer()
			n, err := ser.Read(b)
			if err != nil || n == 0 {
				putBuffer(b)
				log.Printf("LoopSerial: Error reading the serial port. Closing it.")
				ser.Close()
				ser = nil
			} else {
				// Send the character to the channel, but don't wait.
				select {
				case bytesChan <- b[:n]:
					// nothing to do.
					// log.Printf("LoopSerial: Read %d bytes: %s\n", n, string(b))
					// log.Printf("LoopSerial: Read %d bytes.\n", n)

				case <-done:
					// Time to stop.
					ser.Close()
					close(bytesChan)

				default:
				}
			}
		}
	}
}

// Open opens the serial port with name and returns an interface or an error.
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
