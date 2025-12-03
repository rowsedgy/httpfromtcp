package headers

import (
	"bytes"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if h == nil {
		h = make(map[string]string)
	}

	totalConsumed := 0

	for {
		idx := bytes.Index(data[totalConsumed:], []byte("\r\n"))
		if idx == -1 {
			return totalConsumed, false, nil
		}

		line := string(data[totalConsumed : totalConsumed+idx])
		totalConsumed += idx + 2

		if line == "" {
			return totalConsumed, true, nil
		}

		parts := strings.SplitN()
	}
}
