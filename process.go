package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/mdouchement/upathex"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type process struct {
	Name           string
	PaddedName     string
	Hooks          map[string][]string `yaml:"hooks"`
	Pwd            string              `yaml:"pwd"`
	Command        string              `yaml:"command"`
	Environment    map[string]string   `yaml:"environment"`
	Logger         *logger
	LogTrimPattern string `yaml:"log_trim_pattern"`
	IgnoreError    bool   `yaml:"ignore_error"`
	Cancel         context.CancelFunc
	Done           chan struct{}

	mu          sync.Mutex
	waiting     context.Context
	doneWaiting func()
	homedir     string
}

func (p *process) wait() {
	if p.waiting != nil {
		<-p.waiting.Done()
	}
}

func (p *process) update(status map[string][]string) {
	for _, processName := range status["stopped"] {
		for i, name := range p.Hooks["wait"] {
			if name == processName {
				// delete any stopped process from waiting list
				p.Hooks["wait"] = slices.Delete(p.Hooks["wait"], i, i+1)
			}
		}
	}

	if len(p.Hooks["wait"]) == 0 && p.doneWaiting != nil {
		p.doneWaiting()
	}
}

func (p *process) run(ctx context.Context) error {
	logout := p.Logger.WithPrefixName(p.PaddedName).Stdout()
	logerr := p.Logger.WithPrefixName(p.PaddedName).Stderr()

	if p.LogTrimPattern != "" {
		trim, err := regexp.Compile(p.LogTrimPattern)
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

	workdir := p.homedir
	if p.Pwd != "" {
		workdir = p.Pwd
	}

	workdir = upathex.ExpandEnvWithCustom(workdir, p.Environment)

	var err error
	workdir, err = upathex.ExpandTilde(workdir)
	if err != nil {
		return err
	}

	//

	command, err := syntax.NewParser().Parse(strings.NewReader(p.Command), "")
	if err != nil {
		return err
	}

	//
	//

	shell, err := interp.New(
		interp.Dir(workdir),
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

	p.mu.Lock()
	ctx, p.Cancel = context.WithCancel(ctx)
	p.mu.Unlock()

	return shell.Run(ctx, command)
}

func (p *process) wantedDeadOrDead() []string {
	return p.Hooks["kill"]
}

// stop terminates the underlying process whether it is started
func (p *process) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Cancel != nil {
		p.Cancel()
		p.Logger.WithPrefixName(p.PaddedName).Warn("stopped by Composer")
	}
}
