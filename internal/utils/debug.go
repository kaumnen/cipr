package utils

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
)

var debugState = struct {
	sync.Mutex
	enabled bool
	output  io.Writer
}{output: os.Stderr}

// SetDebug enables or disables diagnostic logging. Debug messages always go
// to stderr unless tests replace the writer with setDebugWriter.
func SetDebug(enabled bool) {
	debugState.Lock()
	defer debugState.Unlock()
	debugState.enabled = enabled
}

// Debugf writes a diagnostic message when debug logging is enabled.
func Debugf(format string, args ...any) {
	debugState.Lock()
	defer debugState.Unlock()
	if !debugState.enabled {
		return
	}
	_, _ = fmt.Fprintf(debugState.output, "[debug] "+format+"\n", args...)
}

func setDebugWriter(w io.Writer) func() {
	debugState.Lock()
	previous := debugState.output
	debugState.output = w
	debugState.Unlock()
	return func() {
		debugState.Lock()
		debugState.output = previous
		debugState.Unlock()
	}
}

// SanitizeURL removes URL components that commonly carry credentials or
// tokens before the URL is written to debug output.
func SanitizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid-url>"
	}
	if parsed.User != nil {
		if _, hasPassword := parsed.User.Password(); hasPassword {
			parsed.User = url.UserPassword("xxxxx", "xxxxx")
		} else {
			parsed.User = url.User("xxxxx")
		}
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	return parsed.String()
}
