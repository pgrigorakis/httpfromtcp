package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("error: could not read request - %s", err.Error())
		return &Request{}, err
	}

	requestLine, err := parseRequestLine(string(b))
	if err != nil {
		fmt.Printf("error: could not parse request - %s", err.Error())
		return &Request{}, err
	}

	r := Request{
		RequestLine: *requestLine}

	return &r, nil
}

func parseRequestLine(requestText string) (*RequestLine, error) {
	requestLine := strings.Split(requestText, "\r\n")
	if len(requestLine) == 0 || requestLine[0] == "" {
		return nil, fmt.Errorf("empty request")
	}

	requestLineParts := strings.Fields(requestLine[0])
	if len(requestLineParts) == 0 {
		return nil, fmt.Errorf("empty request line")
	}
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
