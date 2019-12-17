package main

import (
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/sirupsen/logrus"
)

type registry struct {
	observable
	sync.RWMutex
	ready         map[string]*process
	running       map[string]*process
	stopped       map[string]*process
	licenseToKill mapset.Set

	log *logrus.Logger
}

func newRegistry() *registry {
	return &registry{
		ready:         make(map[string]*process),
		running:       make(map[string]*process),
		stopped:       make(map[string]*process),
		licenseToKill: mapset.NewSet(),
	}
}

func (r *registry) register(p *process) {
	r.attachObserver(observerFunc(p.update))

	r.Lock()
	defer r.Unlock()

	r.ready[p.Name] = p
	for _, pname := range p.Hooks["kill"] {
		r.licenseToKill.Add(pname)
	}
}

func (r *registry) updateStatus(p *process, status string) {
	r.Lock()
	switch status {
	case "running":
		delete(r.ready, p.Name)
		r.running[p.Name] = p
	case "stopped":
		delete(r.running, p.Name)
		r.stopped[p.Name] = p
	default:
		panic("Unsupported status") // Should never occur
	}
	r.Unlock()

	r.publish(r.status())
	r.log.WithField("prefix", "registry").Debug(r.status())
}

func (r *registry) status() map[string][]string {
	status := make(map[string][]string)
	r.RLock()
	defer r.RUnlock()

	for name := range r.ready {
		status["ready"] = append(status["ready"], name)
	}
	for name := range r.running {
		status["running"] = append(status["running"], name)
	}
	for name := range r.stopped {
		status["stopped"] = append(status["stopped"], name)
	}
	for _, name := range r.licenseToKill.ToSlice() {
		n := name.(string)
		status["license_to_kill"] = append(status["license_to_kill"], n)
	}

	return status
}

func (r *registry) isAllowedToBeKilled(name string) bool {
	r.RLock()
	defer r.RUnlock()
	return r.licenseToKill.Contains(name)
}

func (r *registry) getProcess(name string) (*process, string) {
	r.RLock()
	defer r.RUnlock()

	if p, ok := r.ready[name]; ok {
		return p, "ready"
	} else if p, ok := r.running[name]; ok {
		return p, "running"
	} else if p, ok := r.stopped[name]; ok {
		return p, "stopped"
	}
	panic("WTF?! Unknown process!")
}

func (r *registry) processes() []*process {
	ps := append(r.readyProcesses(), r.runningProcesses()...)
	return append(ps, r.stoppedProcesses()...)
}

func (r *registry) readyProcesses() []*process {
	r.RLock()
	defer r.RUnlock()

	ps := []*process{}
	for _, process := range r.ready {
		ps = append(ps, process)
	}
	return ps
}

func (r *registry) runningProcesses() []*process {
	r.RLock()
	defer r.RUnlock()

	ps := []*process{}
	for _, process := range r.running {
		ps = append(ps, process)
	}
	return ps
}

func (r *registry) stoppedProcesses() []*process {
	r.RLock()
	defer r.RUnlock()

	ps := []*process{}
	for _, process := range r.stopped {
		ps = append(ps, process)
	}
	return ps
}
