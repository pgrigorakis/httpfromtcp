package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Override(key, value string) {
	h.Remove(key)
	h[strings.ToLower(key)] = value
}

func (h Headers) Remove(key string) {
	for k := range h {
		if strings.EqualFold(k, key) {
			delete(h, k)
		}
	}
}

func (h Headers) Get(key string) (string, bool) {
	for k, value := range h {
		if strings.EqualFold(k, key) {
			return value, true
		}
	}
	return "", false
}

func (h Headers) Set(key, value string) {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			h[k] = strings.Join([]string{
				v,
				value,
			}, ", ")
			return
		}
	}
	h[strings.ToLower(key)] = value
}

var allowedSpecials = map[rune]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'|':  true,
	'~':  true,
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}

	headerText := string(data[:idx])
	header, err := requestHeaderFromString(headerText)
	if err != nil {
		return 0, false, fmt.Errorf("error: could not parse header. %s", err)
	}

	for key, value := range header {
		if _, exists := h[key]; exists {
			newVal := h[key] + ", " + value
			h[key] = newVal
			continue
		}
		h[key] = value
	}
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

	key := strings.ToLower(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		return nil, fmt.Errorf("invalid header")
	}
	if validHeaderKey(key) == false {
		return nil, fmt.Errorf("invalid characters in header")
	}

	return map[string]string{key: value}, nil
}

func validHeaderKey(key string) bool {
	for _, r := range key {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if allowedSpecials[r] {
			continue
		}
		return false
	}
	return true
}
