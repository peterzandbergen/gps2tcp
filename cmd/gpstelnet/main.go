// Main
package main

import (
	"flag"
	"github.com/peterzandbergen/gps2tcp"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
)

var ComPort = flag.String("port", "COM5", "Serial port the GPS Modem is connected to.")
var ListenAddress = flag.String("listen", ":10101", "Listen address and port for tcp.")
var Verbose = flag.Bool("verbose", false, "Display debug messages.")

func main() {
	// Get the command line options.
	flag.Parse()

	if !*Verbose {
		log.SetOutput(ioutil.Discard)
	}

	sig := make(chan os.Signal, 10)

	s := gps2tcp.NewServer(*ListenAddress, *ComPort)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.Run(wg)

	signal.Notify(sig)
	select {
	case ss := <-sig:
		s.Stop()
		log.Printf("Received signal %q.\n", ss)
		log.Println("Waiting for the processes to stop...")
		wg.Wait()
	}
}
