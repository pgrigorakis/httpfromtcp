package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type requestState int

const (
	StateInitialised requestState = iota
	StateDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	req := &Request{
		state: StateInitialised,
	}

	for req.state != StateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				req.state = StateDone
				break
			}
			return nil, err

		}
		readToIndex += bytesRead

		parsedBytes, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[parsedBytes:readToIndex])
		readToIndex -= parsedBytes
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case StateInitialised:
		requestLine, parsedBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if parsedBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = StateDone
		return parsedBytes, nil
	case StateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func parseRequestLine(requestBytes []byte) (*RequestLine, int, error) {
	idx := bytes.Index(requestBytes, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}

	requestText := string(requestBytes[:idx])
	requestLine, err := requestLineFromString(requestText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	requestLineParts := strings.Split(str, " ")
	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("incomplete request")
	}

	method := requestLineParts[0]
	for _, r := range method {
		if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
			return nil, fmt.Errorf("invalid HTTP method: %q", method)
		}
	}

	requestTarget := requestLineParts[1]

	versionParts := strings.Split(requestLineParts[2], "/")

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}
