package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/pgrigorakis/httpfromtcp/internal/request"
	"github.com/pgrigorakis/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		handlerError := &HandlerError{
			StatusCode: response.Status400,
			Message:    "could not parse request",
		}
		handlerError.writeHandlerError(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	handlerError := s.handler(buf, req)

	if handlerError != nil {
		handlerError.writeHandlerError(conn)
		return
	}

	response.WriteStatusLine(conn, response.Status200)
	headers := response.GetDefaultHeaders(buf.Len())
	if err := response.WriteHeaders(conn, headers); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	buf.WriteTo(conn)
}

func (h *HandlerError) writeHandlerError(w io.Writer) error {
	body := h.Message

	err := response.WriteStatusLine(w, h.StatusCode)
	if err != nil {
		return err
	}

	headers := response.GetDefaultHeaders(len(body))
	if err := response.WriteHeaders(w, headers); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	_, err = w.Write([]byte(body))

	return err
}
