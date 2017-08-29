package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type process struct {
	Name           string
	Hooks          map[string][]string `yaml:"hooks"`
	Pwd            string              `yaml:"pwd"`
	Command        string              `yaml:"command"`
	Environment    map[string]string   `yaml:"environment"`
	LogTrimPattern string              `yaml:"log_trim_pattern"`
	IgnoreError    bool                `yaml:"ignore_error"`
	logTrimPattern *regexp.Regexp
	cmd            *exec.Cmd
	Padding        int
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
	if p.LogTrimPattern != "" {
		var err error
		if p.logTrimPattern, err = regexp.Compile(p.LogTrimPattern); err != nil {
			return err
		}
	}

	p.setEnvironment()
	p.cleanCommand()

	args := strings.Split(p.Command, " ")
	p.cmd = exec.Command(args[0], args[1:]...)

	p.setWorkdir()
	p.logStreams()

	return p.cmd.Run()
}

func (p *process) wantedDeadOrDead() []string {
	return p.Hooks["kill"]
}

// kill terminates the underlying process wethwer it is started
func (p *process) kill() {
	// TODO: An inconsistency window can occur between existence check and kill action
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
		log.WithField("prefix", p.paddedName()).Warn("stopped by Composer")
	}
}

// Command preparation
// -------------------

func (p *process) cleanCommand() {
	p.Command = os.ExpandEnv(p.Command)
}

func (p *process) setEnvironment() {
	for k, v := range p.Environment {
		err := os.Setenv(k, os.ExpandEnv(v))
		check(err)
	}
}

func (p *process) setWorkdir() {
	if p.Pwd != "" {
		p.cmd.Dir = p.Pwd
	} else {
		p.cmd.Dir = homedir
	}
}

// Logging
// -------

func (p *process) paddedName() string {
	name := p.Name
	for {
		if len(name) == p.Padding {
			return name
		}
		name = " " + name
	}
}

func (p *process) logStreams() {
	stdout, err := p.cmd.StdoutPipe()
	check(err)

	scout := bufio.NewScanner(stdout)
	scout.Buffer(make([]byte, 4096), cfg.Logger.EntryMaxSize)
	go func() {
		for {
			select {
			case <-p.Done:
				log.WithField("prefix", p.paddedName()).Debug("Stop logging")
				return
			default:
				if scout.Scan() {
					lentries <- &lentry{
						Name:     p.paddedName(),
						Severity: INFO,
						Message:  p.extractMessage(scout.Text()),
					}
				}
			}
		}
	}()

	stderr, err := p.cmd.StderrPipe()
	check(err)

	scerr := bufio.NewScanner(stderr)
	scerr.Buffer(make([]byte, 4096), cfg.Logger.EntryMaxSize)
	go func() {
		for {
			select {
			case <-p.Done:
				log.WithField("prefix", p.paddedName()).Debug("Stop logging")
				return
			default:
				if scerr.Scan() {
					lentries <- &lentry{
						Name:     p.paddedName(),
						Severity: WARN,
						Message:  p.extractMessage(scerr.Text()),
					}
				}
			}
		}
	}()
}

func (p *process) extractMessage(msg string) string {
	if p.logTrimPattern != nil {
		match := p.logTrimPattern.FindStringSubmatch(msg)
		if len(match) > 1 {
			m := MatcherLookup(match, p.logTrimPattern)
			msg = m["message"]
		} else {
			msg = fmt.Sprintf("[!] %s", msg)
		}
	}
	return msg
}
