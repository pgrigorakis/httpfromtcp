package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/pgrigorakis/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	Status200 StatusCode = iota
	Status400
	Status500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case Status200:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case Status400:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case Status500:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	default:
		_, err := w.Write([]byte(""))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%s\r\n", strconv.Itoa(contentLen)),
		"Connection":     "close\r\n",
		"Content-Type":   "text/plain\r\n",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s", key, value)))
		if err != nil {
			return err
		}
	}
	return nil
}
