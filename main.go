// lscom project main.go
package main

import (
	"fmt"
	"github.com/tarm/goserial" // package name is serial
	"io"
	// _ "log"
	"net"
	"os"
	"time"
)

var bla *serial.Config

const (
	ComPort = "COM5"
)

// Create a background version of the program that is started when a connection
// Pass the connection and let the process

func OpenCom(name string) (io.ReadWriteCloser, error) {
	var err error
	for i := 0; i < 5; i++ {
		// First check 5 times if the port exists.
		fmt.Printf("Checking for port presence attempt %d.\n", i+1)
		_, err = os.Stat(name)
		if err != nil {
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	// Try to open the port.
	fmt.Printf("Opening the port.\n")
	cfg := &serial.Config{
		Name:        name,
		Baud:        4800,
		ReadTimeout: 500 * time.Millisecond,
	}
	f, err := serial.OpenPort(cfg)
	if err != nil {
		return nil, err
	}
	//	return &timeOutFile{
	//		ReadWriteCloser: f,
	//		timeout:         10 * time.Second,
	//	}, nil

	return f, nil
}

type timeOutFile struct {
	io.ReadWriteCloser
	timeout time.Duration
}

func (to *timeOutFile) Read(p []byte) (n int, err error) {
	var done = make(chan struct{})

	go func() {
		n, err = to.ReadWriteCloser.Read(p)
		// fmt.Println("Read bytes.")
		done <- struct{}{}
	}()

	select {
	case <-time.Tick(to.timeout):
		n = 0
		err = fmt.Errorf("Timeout reading bytes.")

	case <-done:
		// Bytes have been read, all's well.
	}
	return n, err
}

func main() {

	var n int

	l, err := net.Listen("tcp", ":10101")
	fmt.Printf("Listening on %s\n", l.Addr().String())
	if err != nil {
		fmt.Println("Error starting listener.")
		return
	}
	// Create a channel that keeps a token for the communication channel.
	var token = &struct{}{}
	var runningChan = make(chan *struct{}, 1)
	// Put the token on the channel.
	runningChan <- token
	token = nil
	// I have the token.
	for {

		fmt.Printf("Waiting for connection.\n")
		conn, err := l.Accept()
		if err != nil {
			return
		}

		select {
		case token = <-runningChan:
		default:
		}

		if token == nil {
			conn.Close()
		} else {
			fmt.Printf("starting connection for %s\n", conn.RemoteAddr().String())
			// Get the GPS com reader.
			fmt.Printf("Opening port: %s\n", ComPort)
			f, err := OpenCom(ComPort)
			if err != nil {
				// Failed to get the port, close the connection.
				fmt.Printf("Error opening the comport: %s\n", err.Error())
				conn.Close()
			} else {
				fmt.Printf("Serving GPS Data.\n")
				go func(c net.Conn, r io.ReadCloser) {
					// Wait for the token before starting.
					token := <-runningChan
					defer c.Close()
					defer r.Close()
					defer func(c chan *struct{}) {
						c <- token
					}(runningChan)
					var buf [1]byte
					b := buf[:]
					fmt.Printf("Entering serving loop.\n")
					for {
						n, err = r.Read(b)
						if err != nil {
							fmt.Printf("Error occured reading byte: %s\n", err.Error())
							return
						}
						n, err = c.Write(b)
						if err != nil {
							fmt.Println("Failed writing to socket: %s", err.Error())
							return
						}
					}
				}(conn, f)
				// Start the process.
				runningChan <- token
				token = nil
			}
		}
	}
}

//func main() {
//	for {
//		fmt.Println("Opening com...")
//		f, err := OpenCom(ComPort)
//		if err != nil {
//			fmt.Printf("Error opening: %s\n", err.Error())
//			time.Sleep(time.Second)
//		} else {
//			// Opened, start reading.
//			fmt.Printf("Opened. Trying again after 5 seconds.\n")

//			tor := &timeOutFile{
//				ReadWriteCloser: f,
//				timeout:         5 * time.Second,
//			}
//			var b = make([]byte, 1)
//			for {
//				n, err := tor.Read(b)
//				if err != nil {
//					fmt.Printf("Error reading byte: %s\n", err.Error())
//					break
//				}
//				fmt.Print(string(b[0:n]))
//			}
//			fmt.Println("Closing connection.")
//			tor.Close()
//			fmt.Println("Connection closed.")
//			time.Sleep(5 * time.Second)
//		}
//	}

//	var n int

//	l, err := net.Listen("tcp", ":10101")
//	if err != nil {
//		fmt.Println("Error starting listener.")
//		return
//	}
//	for {
//		conn, err := l.Accept()
//		if err != nil {
//			return
//		}

//		fmt.Println("starting connection")
//		go func(c net.Conn) {
//			defer c.Close()
//			var buf [1]byte
//			b := buf[:]
//			for {
//				n, err = f.Read(b)
//				if err != nil {
//					fmt.Printf("Error occured: %s\n", err.Error())
//					break
//				}
//				n, err = c.Write(b)
//				if err != nil {
//					fmt.Println("Failed writing to socket: %s", err.Error())
//					return
//				}
//			}
//		}(conn)

//	}

//}
