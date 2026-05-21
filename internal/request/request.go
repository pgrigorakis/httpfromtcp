package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type InitState int

const (
	StateInitialised InitState = iota
	StateDone
)

type Request struct {
	RequestLine RequestLine
	InitState   int
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

	r := Request{
		RequestLine: RequestLine{},
		InitState:   int(StateInitialised),
	}

	for r.InitState != int(StateDone) {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			r.InitState = int(StateDone)
			break
		}

		readToIndex += n

		parsedBytes, err := r.parse(buf[:readToIndex])
		if err != nil {
			return &r, err
		}
		copy(buf, buf[parsedBytes:readToIndex])
		readToIndex -= parsedBytes
	}
	return &r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.InitState == int(StateInitialised) {
		requestLine, parsedBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if parsedBytes == 0 {
			return 0, nil
		} else {
			r.RequestLine = *requestLine
			r.InitState = int(StateDone)
			return parsedBytes, nil
		}
	} else if r.InitState == int(StateDone) {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	} else {
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
