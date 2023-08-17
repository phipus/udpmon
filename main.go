package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type service interface {
	Run(done <-chan struct{})
}

func main() {
	var (
		listenAddress    = flag.String("listen", "", "specify an address to listen on")
		serverAddress    = flag.String("server", "", "specify the destination server")
		logFile          = flag.String("logfile", "", "specify a file to write the logs to")
		frequency        = flag.Int("frequency", 100, "specify the amount of time between requests in millisecons")
		timeout          = flag.Int("timeout", 100, "specify the timeout for round trips in milliseconds")
		latencyLogFile   = flag.String("latencylogfile", "", "specify a file to store latency logs")
		latencyThreshold = flag.Int("latencythreshold", 80, "specify the latency threshold starting from which it is logged in the latency logfile")
	)

	flag.Parse()

	var outFile io.Writer
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file %s: %v\n", *logFile, err)
			os.Exit(1)
		}
		defer f.Close()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	var latencyLog io.Writer
	if *latencyLogFile != "" {
		f, err := os.OpenFile(*latencyLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open latency log file %s: %v\n", *latencyLogFile, err)
			os.Exit(1)
		}
		defer f.Close()
		latencyLog = f
	}

	var svc service
	if *listenAddress != "" {
		s, err := NewServer(*listenAddress, outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start server: %v", err)
			os.Exit(1)
		}
		defer s.Close()
		svc = s
	} else if *serverAddress != "" {
		c, err := NewClient(*serverAddress, &ClientOptions{
			Timeout:          time.Duration(*timeout) * time.Millisecond,
			Frequency:        time.Duration(*frequency) * time.Millisecond,
			LatencyThreshold: time.Duration(*latencyThreshold) * time.Millisecond,
			LogFile:          outFile,
			LatencyLogFile:   latencyLog,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start client: %v", err)
			os.Exit(1)
		}
		defer c.Close()
		svc = c
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}

	done := make(chan struct{})
	serviceDone := make(chan struct{})

	go func() {
		defer close(serviceDone)
		svc.Run(done)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-serviceDone:
			return
		case <-sigChan:
			close(done)
		}
	}

}
