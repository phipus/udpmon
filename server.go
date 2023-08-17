package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Server struct {
	conn    net.PacketConn
	addr    string
	logFile io.Writer
}

func NewServer(addr string, logFile io.Writer) (*Server, error) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on address %s: %w", addr, err)
	}

	return &Server{
		conn:    conn,
		addr:    addr,
		logFile: logFile,
	}, nil
}

func (s *Server) Close() error {
	return s.conn.Close()
}

func (s *Server) Run(done <-chan struct{}) {
	fmt.Fprintf(s.logFile, "at %s server started listening on %s\n", time.Now().Format(time.RFC1123), s.addr)
	defer func() {
		fmt.Fprintf(s.logFile, "at %s server stopped listening on %s\n", time.Now().Format(time.RFC1123), s.addr)
	}()

	buf := make([]byte, 1024)

	for {
		select {
		case <-done:
			return
		default:
			// continue
		}

		s.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, from, err := s.conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			fmt.Fprintf(s.logFile, "error reading from %s: %v\n", from, err)
			continue
		}
		// fmt.Printf("read %s from %s succeeded\n", string(buf[0:n]), from)
		_, err = s.conn.WriteTo(buf[:n], from)
		if err != nil {
			fmt.Fprintf(s.logFile, "error writing to %s: %v\n", from, err)
		}
	}
}
