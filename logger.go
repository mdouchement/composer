package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"runtime"
)

// ANSI color
type Color string

const (
	Reset  Color = "\033[0m" // \033 == \x1B
	Red    Color = "\033[31m"
	Yellow Color = "\033[33m"
	Cyan   Color = "\033[36m"
)

// ----------------
// -------------
// Logging
// -----
// ---

type logger struct {
	w      io.Writer
	prefix string
}

func (l *logger) WithPrefixName(name string) *logger {
	name += ": " // space important
	if l.prefix != "" {
		name = l.prefix + name
	}

	return &logger{
		w:      l.w,
		prefix: name,
	}
}

func (l *logger) Info(args ...any) {
	l.println(Cyan, args...)
}

func (l *logger) Warn(args ...any) {
	l.println(Yellow, args...)
}

func (l *logger) Error(args ...any) {
	l.println(Red, args...)
}

func (l *logger) println(c Color, args ...any) {
	buf := bytes.NewBuffer(make([]byte, len(l.prefix)+len(args)))

	if l.prefix != "" {
		fmt.Fprintf(buf, "%s%s%s", c, l.prefix, Reset) //nolint: errcheck
	}

	fmt.Fprint(buf, args...) //nolint: errcheck
	buf.WriteString("\n")    //nolint: errcheck

	buf.WriteTo(l.w) //nolint: errcheck
}

func (l *logger) Stdout() *std {
	return l.writer(Cyan)
}

func (l *logger) Stderr() *std {
	return l.writer(Yellow)
}

func (l *logger) writer(c Color) *std {
	var prefix []byte
	if l.prefix != "" {
		prefix = fmt.Appendf(nil, "%s%s%s", c, l.prefix, Reset)
	}
	s := newStd(l.w, prefix)

	return s
}

//
//

type std struct {
	io.Writer
	prefix []byte
	w      io.Writer
	trim   *regexp.Regexp
}

func newStd(w io.Writer, prefix []byte) *std {
	reader, writer := io.Pipe()

	s := &std{
		Writer: writer,
		w:      w,
		prefix: prefix,
	}

	// Start a new goroutine to scan the input and write it to the logger using the specified print function.
	// It splits the input into chunks of up to 64KB to avoid buffer overflows.
	go s.writerScanner(reader)

	// Set a finalizer function to close the writer when it is garbage collected
	runtime.SetFinalizer(writer, writerFinalizer)

	return s
}

func (s *std) extractMessage(msg []byte) []byte {
	if s.trim == nil {
		return msg
	}

	match := s.trim.FindSubmatch(msg)
	if len(match) < 2 {
		// return append([]byte("[!] "), msg...)
		return msg
	}

	m := msg[:0]
	for _, p := range match[1:] {
		m = append(m, p...)
	}

	return m
}

func (s *std) writerScanner(reader *io.PipeReader) {
	scanner := bufio.NewScanner(reader)

	// Set the buffer size to the maximum token size to avoid buffer overflows
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), bufio.MaxScanTokenSize)

	// Define a split function to split the input into chunks of up to 64KB
	chunkSize := bufio.MaxScanTokenSize - len(s.prefix) // ~64KB
	splitFunc := func(data []byte, atEOF bool) (int, []byte, error) {
		if len(data) >= chunkSize {
			return chunkSize, data[:chunkSize], nil
		}

		return bufio.ScanLines(data, atEOF)
	}

	// Use the custom split function to split the input
	scanner.Split(splitFunc)

	// Scan the input and write it to the logger using the specified print function
	for scanner.Scan() {
		p := bytes.TrimRight(scanner.Bytes(), "\r\n")
		p = s.extractMessage(p)
		if len(s.prefix) != 0 {
			p = append(s.prefix, p...)
		}
		p = append(p, '\n')

		s.w.Write(p) //nolint: errcheck
	}

	// If there was an error while scanning the input, log an error
	if err := scanner.Err(); err != nil {
		io.WriteString(s.w, fmt.Sprintf("Error while reading from Writer: %s\n", err)) //nolint: errcheck
	}

	// Close the reader when we are done
	reader.Close() //nolint: errcheck
}

func writerFinalizer(writer *io.PipeWriter) {
	writer.Close() //nolint: errcheck
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
