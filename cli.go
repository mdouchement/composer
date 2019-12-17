package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func command(log *logrus.Logger, homedir string) *cobra.Command {
	var config string
	parser := &parser{
		log:     log,
		homedir: homedir,
	}

	command := &cobra.Command{
		Use:   "start",
		Short: "Start all processes",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) (err error) {
			registry, err := parser.parseConfig(config)
			if err != nil {
				return err
			}

			runner := &processor{
				log:       log,
				reg:       registry,
				terminate: make(chan []string, len(registry.processes())),
				errors:    make(chan error, len(registry.processes())),
			}

			//

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)
			go func() {
				<-signals
				log.Info("Gracefully shutdown composer")

				runner.termination = true
				for _, process := range registry.processes() {
					runner.stop(process)
				}
			}()

			//

			return runner.perform()
		},
	}
	command.Flags().StringVarP(&config, "config", "c", "", "Configuration file")

	return command
}

// ----------------
// -------------
// Config
// -----
// ---

type parser struct {
	log     *logrus.Logger
	homedir string
}

func (ps *parser) parseConfig(path string) (*registry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	raw := make(map[string]interface{})
	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	// No longer used for the moment
	//
	// cfg, err := ps.parseSettings(raw["settings"])
	// if err != nil {
	// 	return nil, err
	// }

	reg := newRegistry()
	reg.log = ps.log
	services, err := ps.parseServices(raw["services"])
	if err != nil {
		return nil, err
	}

	for name, p := range services {
		p.Name = name
		p.Done = make(chan struct{})
		p.Logger = logrus.NewEntry(ps.log)
		p.homedir = ps.homedir
		reg.register(p)
	}

	return reg, nil
}

func (ps *parser) parseSettings(value interface{}) (*configuration, error) {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize settings")
	}

	// defaults
	cfg := configuration{
		Logger: logConfiguration{
			BufferSize:   42,
			EntryMaxSize: bufio.MaxScanTokenSize,
		},
	}

	err = yaml.Unmarshal(raw, &cfg)
	return &cfg, errors.Wrap(err, "could not parse settings")
}

func (ps *parser) parseServices(value interface{}) (map[string]*process, error) {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize services")
	}

	services := make(map[string]*process)
	err = yaml.Unmarshal(raw, &services)
	return services, errors.Wrap(err, "could not parse services")
}
