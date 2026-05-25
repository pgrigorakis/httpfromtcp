package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/pgrigorakis/httpfromtcp/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, fmt.Errorf("could not create listener to port %d", port)
	}
	server := &Server{listener: listener}

	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	err := s.listener.Close()
	if err != nil {
		return fmt.Errorf("could not close listener")
	}
	s.closed.Store(true)
	return nil
}

func (s *Server) listen() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return fmt.Errorf("server closed")
			}
			return fmt.Errorf("could not open connection")
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	response.WriteStatusLine(conn, response.Status200)
	headers := response.GetDefaultHeaders(0)
	response.WriteHeaders(conn, headers)

	fmt.Fprintf(conn, "\nConnection to %s closed", conn.RemoteAddr())
}
