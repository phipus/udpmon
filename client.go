package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type ClientOptions struct {
	Timeout          time.Duration
	Frequency        time.Duration
	LatencyThreshold time.Duration
	LogFile          io.Writer
	LatencyLogFile   io.Writer
}

type Client struct {
	conn       net.Conn
	serverAddr string
	opts       ClientOptions
}

func NewClient(serverAddr string, opts *ClientOptions) (*Client, error) {
	conn, err := net.Dial("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       conn,
		serverAddr: serverAddr,
		opts:       *opts,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Run(done <-chan struct{}) {
	fmt.Fprintf(c.opts.LogFile, "at %s client started connecting to %s\n", time.Now().Format(time.RFC1123), c.serverAddr)
	defer func() {
		fmt.Fprintf(c.opts.LogFile, "at %s client stopped connecting to %s\n", time.Now().Format(time.RFC1123), c.serverAddr)
	}()

	buf := make([]byte, 1024)
	ticker := time.NewTicker(c.opts.Frequency)

	printErrors := true

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// continue
		}

		t := time.Now()
		tString := t.String()

		_, err := c.conn.Write([]byte(tString))
		if err != nil {
			if printErrors {
				fmt.Fprintf(c.opts.LogFile, "write %s to server failed: %v\n", tString, err)
				printErrors = false
			}
			continue
		}

		c.conn.SetReadDeadline(t.Add(c.opts.Timeout))
		n, err := c.conn.Read(buf)
		if err != nil || string(buf[:n]) != tString {
			if err == nil {
				err = fmt.Errorf("the returned value %s did not match the sent one", string(buf[:n]))
			}
			if printErrors {
				fmt.Fprintf(c.opts.LogFile, "read %s from server failed: %v\n", tString, err)
				printErrors = false
			}
			continue
		}

		if !printErrors {
			fmt.Fprintf(c.opts.LogFile, "send/read %s succeeded: error condition resolved\n", tString)
			printErrors = true
		}

		c.printLatency(t, time.Since(t))

	}
}

func (c *Client) printLatency(t time.Time, latency time.Duration) {
	if c.opts.LatencyLogFile != nil && latency >= c.opts.LatencyThreshold {
		fmt.Fprintf(c.opts.LatencyLogFile, "latency at %s was %d ms\n", t.Format(time.RFC1123), latency.Milliseconds())
	}
}
