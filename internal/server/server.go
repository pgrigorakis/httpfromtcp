package server

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/pgrigorakis/httpfromtcp/internal/request"
	"github.com/pgrigorakis/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	server := &Server{listener: listener, handler: handler}

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
	rw := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		body := []byte(fmt.Sprintf("error parsing request: %v", err))
		rw.WriteResponse(response.Status400, body)
		return
	}
	s.handler(rw, req)
}
