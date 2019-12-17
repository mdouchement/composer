package main

import (
	"sync"
)

var (
	errors    chan error
	terminate chan []string
)

func perform(reg *registry) {
	padding := getPadding(reg)
	terminate = make(chan []string, len(reg.processes()))
	errors = make(chan error, len(reg.processes()))
	go terminator(reg)
	go handleErrors(reg)

	var n sync.WaitGroup
	for _, p := range reg.readyProcesses() {
		p.Padding = padding

		n.Add(1)
		go func(p *process) {
			defer n.Done()
			p.wait()
			reg.updateStatus(p, "running")
			err := p.run()
			if !p.IgnoreError && err != nil && !(err.Error() == "signal: killed" && reg.isAllowedToBeKilled(p.Name)) {
				errors <- err
			}
			close(p.Done)
			reg.updateStatus(p, "stopped")
			terminate <- p.wantedDeadOrDead()
		}(p)
	}
	n.Wait()
}

func handleErrors(reg *registry) {
	for err := range errors {
		if err != nil {
			log.WithField("prefix", "processor").Errorf("%s ; %#v", err.Error(), err)
			for _, p := range reg.runningProcesses() {
				stop(reg, p)
			}
			fail(err.Error())
		}
	}
}

func terminator(reg *registry) {
	for names := range terminate {
		stopAllGivenNames(reg, names)
	}
}

func stopAllGivenNames(reg *registry, names []string) {
	for _, name := range names {
		log.WithField("prefix", "processor").Warn(name)
		p, status := reg.getProcess(name)
		switch status {
		case "ready":
			reg.updateStatus(p, "stopped")
		case "running":
			stop(reg, p)
		case "stopped":
			// nothing to do here
		}
	}
}

func stop(reg *registry, p *process) {
	p.stop()
	reg.updateStatus(p, "stopped")
}

func getPadding(reg *registry) int {
	var length int
	for _, process := range reg.processes() {
		if l := len(process.Name); l > length {
			length = l
		}
	}
	return length
}
