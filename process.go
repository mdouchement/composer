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
	Reload         []string            `yaml:"reload"`
	Environment    map[string]string   `yaml:"environment"`
	Logger         *logrus.Entry
	LogTrimPattern string `yaml:"log_trim_pattern"`
	IgnoreError    bool   `yaml:"ignore_error"`
	Padding        int
	Cancel         context.CancelFunc
	Done           chan struct{}

	homedir string
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

	workdir := p.homedir
	if p.Pwd != "" {
		workdir = p.Pwd
	}

	workdir = os.Expand(workdir, func(k string) string {
		if e, ok := p.Environment[k]; ok {
			return e
		}
		return fmt.Sprintf("${%s}", k)
	})
	workdir = os.ExpandEnv(workdir)

	//

	command, err := syntax.NewParser().Parse(strings.NewReader(p.Command), "")
	if err != nil {
		return err
	}

	//

	watcher, err := newWatcher(workdir, p.Reload)
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

	ctx := context.Background()
	ctx, p.Cancel = context.WithCancel(ctx)

	//

	errCh := make(chan error)
	run := func(ctx context.Context) {
		go func() {
			errCh <- shell.Run(ctx, command)
		}()
	}

	current, reload := context.WithCancel(ctx)
	run(current)
	for {

		select {
		case err = <-errCh:
			reload() // The stop already spreaded by parrent context (make go vet happy)
			return err
		case <-watcher.watch():
			p.Logger.WithField("prefix", "composer").Infof("Reloading %s", p.Name)
			reload()
			current, reload = context.WithCancel(ctx)
			run(current)
		case <-ctx.Done():
			reload() // The stop already spreaded by parrent context (make go vet happy)
			return nil
		}
	}
}

func (p *process) wantedDeadOrDead() []string {
	return p.Hooks["kill"]
}

// stop terminates the underlying process whether it is started
func (p *process) stop() {
	// TODO: An inconsistency window can occur between existence check and kill action
	if p.Cancel != nil {
		p.Cancel()
		p.Logger.WithField("prefix", p.paddedName()).Warn("stopped by Composer")
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
