package main

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type processor struct {
	log         *logrus.Logger
	errors      chan error
	terminate   chan []string
	reg         *registry
	termination bool
}

func (p *processor) perform() error {
	go p.terminator()
	go p.handleErrors()

	padding := p.getPadding()
	var n sync.WaitGroup

	for _, proc := range p.reg.readyProcesses() {
		proc.Padding = padding

		n.Add(1)
		go func(proc *process) {
			defer n.Done()
			proc.wait()
			p.reg.updateStatus(proc, "running")
			err := proc.run()
			if !proc.IgnoreError && err != nil && !p.reg.isAllowedToBeKilled(proc.Name) {
				p.errors <- err
			}
			close(proc.Done)
			p.reg.updateStatus(proc, "stopped")
			p.terminate <- proc.wantedDeadOrDead()
		}(proc)
	}
	n.Wait()

	return nil
}

func (p *processor) handleErrors() {
	for err := range p.errors {
		if err != nil {
			if p.termination {
				continue // Already in termination state
			}
			p.termination = true

			p.log.WithField("prefix", "processor").Errorf("%s ; %#v", err.Error(), err)
			for _, process := range p.reg.runningProcesses() {
				p.stop(process)
			}
		}
	}
}

func (p *processor) terminator() {
	for names := range p.terminate {
		p.stopAllGivenNames(names)
	}
}

func (p *processor) stopAllGivenNames(names []string) {
	for _, name := range names {
		p.log.WithField("prefix", "processor").Warn(name)
		process, status := p.reg.getProcess(name)
		switch status {
		case "ready":
			p.reg.updateStatus(process, "stopped")
		case "running":
			p.stop(process)
		case "stopped":
			// nothing to do here
		}
	}
}

func (p *processor) stop(process *process) {
	process.stop()
	p.reg.updateStatus(process, "stopped")
}

func (p *processor) getPadding() int {
	var length int
	for _, process := range p.reg.processes() {
		if l := len(process.Name); l > length {
			length = l
		}
	}
	return length
}
