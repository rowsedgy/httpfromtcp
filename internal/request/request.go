package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	State       State
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type State int

const (
	initialized State = iota
	done
)

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		State: initialized,
	}

	for req.State != done {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.State = done
				break
			}
			return nil, err
		}
		readToIndex += n

		bytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[bytesParsed:])
		readToIndex -= bytesParsed
	}
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineString := string(data[:idx])
	requestLine, err := parseRequestLineString(requestLineString)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func parseRequestLineString(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, errors.New("malformed request")
	}

	method := parts[0]
	target := parts[1]
	version := parts[2]

	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	if !strings.HasPrefix(target, "/") {
		return nil, errors.New("wrong target format")
	}

	versionParts := strings.Split(version, "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("wrong version formatting: %s", version)
	}

	httpPart := versionParts[0]
	versionPart := versionParts[1]

	if httpPart != "HTTP" || versionPart != "1.1" {
		return nil, errors.New("wrong version format")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   versionPart,
	}, nil

}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case initialized:
		reqLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.State = done
		return n, nil
	case done:
		return 0, fmt.Errorf("trying to return data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
