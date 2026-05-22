package headers

import (
	"bytes"
	"fmt"
	"maps"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 0, true, nil
	}

	headerText := string(data[:idx])
	header, err := requestHeaderFromString(headerText)
	if err != nil {
		return 0, false, fmt.Errorf("error: could not parse header.")
	}

	maps.Copy(h, header)
	return idx + 2, false, nil
}

func requestHeaderFromString(str string) (map[string]string, error) {
	if len(str) == 0 || str[0] == ' ' || str[0] == '\t' {
		return nil, fmt.Errorf("invalid header spacing")
	}

	parts := strings.SplitN(str, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid header")
	}

	key := parts[0]
	value := strings.TrimSpace(parts[1])

	if key == "" || value == "0" {
		return nil, fmt.Errorf("invalid header")
	}

	return map[string]string{key: value}, nil
}
