package main

import prefixer "github.com/x-cray/logrus-prefixed-formatter"

const (
	ERROR = iota
	WARN
	INFO
	DEBUG
)

type lentry struct {
	Name     string
	Severity int
	Message  string
}

var lentries chan *lentry

func startLogger() {
	log.Formatter = new(prefixer.TextFormatter)

	// Processes logger initialization
	lentries = make(chan *lentry, cfg.Logger.BufferSize)
	go func() {
		for entry := range lentries {
			switch entry.Severity {
			case DEBUG:
				log.WithField("prefix", entry.Name).Debug(entry.Message)
			case INFO:
				log.WithField("prefix", entry.Name).Info(entry.Message)
			case WARN:
				log.WithField("prefix", entry.Name).Warn(entry.Message)
			case ERROR:
				log.WithField("prefix", entry.Name).Error(entry.Message)
			}
		}
	}()
}
