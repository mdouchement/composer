package main

import (
	"io"

	"github.com/mdouchement/nregexp"
)

// ----------------
// -------------
// Logging
// -----
// ---

type logger struct {
	w    io.Writer
	trim *nregexp.NRegexp
}

func (l *logger) Write(p []byte) (int, error) {
	p = l.extractMessage(p)
	return l.w.Write(p)
}

func (l *logger) extractMessage(msg []byte) []byte {
	if l.trim != nil {
		match := l.trim.NamedCapture(msg)
		if m, ok := match["message"]; ok {
			return m
		}

		return append([]byte("[!] "), msg...)
	}

	return msg
}

// ----------------
// -------------
// /dev/null
// -----
// ---

type devNull struct{}

func (devNull) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func (devNull) Close() error {
	return nil
}
