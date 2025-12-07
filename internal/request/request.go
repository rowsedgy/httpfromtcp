package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       State
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type State int

const (
	requestStateInitialized State = iota
	requestStateDone
	requestStateParsingHeaders
)

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		State:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}

	for req.State != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.State != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.State, n)
				}

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
	totalBytesParsed := 0
	for r.State != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}

	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, nil
		}
		if done {
			r.State = requestStateDone
		}
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
