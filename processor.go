package main

import (
	"context"
	"fmt"
	"sync"
)

type processor struct {
	log       *logger
	errors    chan error
	terminate chan []string
	reg       *registry

	m           sync.Mutex
	termination bool
}

func (p *processor) perform(ctx context.Context) {
	go p.terminator()
	go p.handleErrors()

	template := fmt.Sprintf("%%%ds", p.getPadding())
	var n sync.WaitGroup

	for _, proc := range p.reg.readyProcesses() {
		proc.PaddedName = fmt.Sprintf(template, proc.Name)

		n.Add(1)
		go func(proc *process) {
			defer n.Done()

			proc.wait()
			p.reg.updateStatus(proc, "running")
			err := proc.run(ctx)
			if !proc.IgnoreError && err != nil && !p.reg.isAllowedToBeKilled(proc.Name) {
				p.errors <- err
			}
			close(proc.Done)
			p.reg.updateStatus(proc, "stopped")
			p.terminate <- proc.wantedDeadOrDead()
		}(proc)
	}

	n.Wait()
}

func (p *processor) handleErrors() {
	var termination bool

	for err := range p.errors {
		if err == nil {
			continue
		}

		if termination {
			continue
		}

		if err != context.Canceled {
			p.log.WithPrefixName("processor").Error(fmt.Sprintf("%s ; %#v", err.Error(), err))
		}

		p.shutdown()
		termination = true
	}
}

func (p *processor) terminator() {
	for names := range p.terminate {
		p.stopAllGivenNames(names)
	}
}

func (p *processor) stopAllGivenNames(names []string) {
	for _, name := range names {
		p.log.WithPrefixName("processor").Warn(name)
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

func (p *processor) shutdown() {
	p.m.Lock()
	defer p.m.Unlock()

	if p.termination {
		return // Already in termination state
	}

	p.log.WithPrefixName("processor").Info("Gracefully shutdown composer")
	p.termination = true
	defer p.reg.shutdown()

	for _, process := range p.reg.readyProcesses() {
		p.stop(process)
	}

	for _, process := range p.reg.runningProcesses() {
		p.stop(process)
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
