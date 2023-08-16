package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
)

var (
	listenAddress = flag.String("listen", "", "specify an address to listen on")
	serverAddress = flag.String("server", "", "specify the destination server")
	logFile       = flag.String("logfile", "", "specify a file to write the logs to")
	timeout       = flag.Int("timeout", 100, "specify the timeout for round trips in milliseconds")
)

func main() {
	flag.Parse()

	var outFile io.Writer
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file %s: %v", *logFile, err)
		}
		defer f.Close()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	done := make(chan struct{})
	mainDone := make(chan struct{})

	if *listenAddress != "" {
		go func() {
			defer close(mainDone)

			err := runServer(*listenAddress, outFile, done)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}()

	} else if *serverAddress != "" {
		go func() {
			defer close(mainDone)

			err := runClient(*serverAddress, outFile, done)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}()

	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case <-mainDone:
			return
		case <-sigChan:
			close(done)
		}
	}

}
