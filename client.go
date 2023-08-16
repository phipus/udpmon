package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func runClient(serverAddr string, logFile io.Writer, done chan struct{}) error {
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	fmt.Fprintf(logFile, "at %s client started connecting to %s\n", time.Now().Format(time.RFC1123), serverAddr)
	defer func() {
		fmt.Fprintf(logFile, "at %s client stopped connecting to %s\n", time.Now().Format(time.RFC1123), serverAddr)
	}()

	buf := make([]byte, 1024)
	ticker := time.NewTicker(time.Duration(*timeout) * time.Millisecond)

	printErrors := true

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			// continue
		}

		t := time.Now()
		tString := t.String()

		_, err = conn.Write([]byte(tString))
		if err != nil {
			fmt.Fprintf(logFile, "write %s to server failed: %v\n", tString, err)
			continue
		}

		conn.SetReadDeadline(t.Add(time.Duration(*timeout) * time.Millisecond))
		n, err := conn.Read(buf)
		if err != nil || string(buf[:n]) != tString {
			if err == nil {
				err = fmt.Errorf("the returned value %s did not match the sent one", string(buf[:n]))
			}
			if printErrors {
				fmt.Fprintf(logFile, "read %s from server failed: %v\n", tString, err)
				printErrors = false
			}
			continue
		}

		if !printErrors {
			fmt.Fprintf(logFile, "send/read %s succeeded: error condition restored\n", tString)
			printErrors = true
		}
	}
}
