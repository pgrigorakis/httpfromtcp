package response

import (
	"fmt"
	"io"

	"github.com/pgrigorakis/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	Status200 StatusCode = 200
	Status400 StatusCode = 400
	Status500 StatusCode = 500
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

type Writer struct {
	writer      io.Writer
	writerState writerState
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer:      writer,
		writerState: writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}

	switch statusCode {
	case Status200:
		_, err := w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case Status400:
		_, err := w.writer.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case Status500:
		_, err := w.writer.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}
	default:
		_, err := w.writer.Write([]byte(""))
		if err != nil {
			return err
		}
	}
	w.writerState = writerStateHeaders
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%d", contentLen),
		"Connection":     "close",
		"Content-Type":   "text/html",
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	for key, value := range headers {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.writerState = writerStateBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	_, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *Writer) WriteResponse(statusCode StatusCode, p []byte) error {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		return err
	}

	headers := GetDefaultHeaders(len(p))
	err = w.WriteHeaders(headers)
	if err != nil {
		return err
	}

	_, err = w.WriteBody(p)
	if err != nil {
		return err
	}

	return err
}
