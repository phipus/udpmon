package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func runServer(addr string, logFile io.Writer, done chan struct{}) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("udpmon: failed to listen on address %s: %w", addr, err)
	}
	defer conn.Close()

	fmt.Fprintf(logFile, "at %s server started listening on %s\n", time.Now().Format(time.RFC1123), addr)
	defer func() {
		fmt.Fprintf(logFile, "at %s server stopped listening on %s\n", time.Now().Format(time.RFC1123), addr)
	}()

	buf := make([]byte, 1024)

	for {
		select {
		case <-done:
			return nil
		default:
			// continue
		}

		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			fmt.Fprintf(logFile, "error reading from %s: %v\n", from, err)
			continue
		}
		// fmt.Printf("read %s from %s succeeded\n", string(buf[0:n]), from)
		_, err = conn.WriteTo(buf[:n], from)
		if err != nil {
			fmt.Fprintf(logFile, "error writing to %s: %v\n", from, err)
		}
	}

}
