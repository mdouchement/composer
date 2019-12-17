package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mdouchement/nregexp"
	"github.com/sirupsen/logrus"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type process struct {
	Name           string
	Hooks          map[string][]string `yaml:"hooks"`
	Pwd            string              `yaml:"pwd"`
	Command        string              `yaml:"command"`
	Environment    map[string]string   `yaml:"environment"`
	Logger         *logrus.Entry
	LogTrimPattern string `yaml:"log_trim_pattern"`
	IgnoreError    bool   `yaml:"ignore_error"`
	Padding        int
	Cancel         context.CancelFunc
	Done           chan struct{}
}

// TODO: try to block without using loop + sleep
func (p *process) wait() {
	for len(p.Hooks["wait"]) != 0 {
		time.Sleep(1 * time.Second)
	}
}

func (p *process) update(status map[string][]string) {
	for _, processName := range status["stopped"] {
		for i, name := range p.Hooks["wait"] {
			if name == processName {
				// delete any stopped process from waiting list
				p.Hooks["wait"] = append(p.Hooks["wait"][:i], p.Hooks["wait"][i+1:]...)
			}
		}
	}
}

func (p *process) run() error {
	logout := &logger{
		w: p.Logger.WithField("prefix", p.paddedName()).WriterLevel(logrus.InfoLevel),
	}
	logerr := &logger{
		w: p.Logger.WithField("prefix", p.paddedName()).WriterLevel(logrus.WarnLevel),
	}

	if p.LogTrimPattern != "" {
		trim, err := nregexp.Compile(p.LogTrimPattern)
		if err != nil {
			return err
		}
		logout.trim = trim
		logerr.trim = trim
	}

	//

	environ := os.Environ()
	for k, v := range p.Environment {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}

	//

	workdir := interp.Dir(homedir)
	if p.Pwd != "" {
		workdir = interp.Dir(p.Pwd)
	}

	//

	command, err := syntax.NewParser().Parse(strings.NewReader(p.Command), "")
	if err != nil {
		return err
	}

	//
	//

	shell, err := interp.New(
		workdir,
		interp.Env(expand.ListEnviron(environ...)),

		interp.OpenHandler(func(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			if path == "/dev/null" {
				return devNull{}, nil
			}
			return interp.DefaultOpenHandler()(ctx, path, flag, perm)
		}),

		interp.StdIO(os.Stdin, logout, logerr),
	)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, p.Cancel = context.WithCancel(ctx)
	return shell.Run(ctx, command)
}

func (p *process) wantedDeadOrDead() []string {
	return p.Hooks["kill"]
}

// stop terminates the underlying process whether it is started
func (p *process) stop() {
	// TODO: An inconsistency window can occur between existence check and kill action
	if p.Cancel != nil {
		p.Cancel()
		log.WithField("prefix", p.paddedName()).Warn("stopped by Composer")
	}
}

func (p *process) paddedName() string {
	name := p.Name
	for {
		if len(name) == p.Padding {
			return name
		}
		name = " " + name
	}
}

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

// https://github.com/go-task/task/blob/master/internal/execext/devnull.go

var _ io.ReadWriteCloser = devNull{}

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
