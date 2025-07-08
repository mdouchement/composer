package main

import (
	"context"
	"io"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func command(log *logger, homedir string) *cobra.Command {
	var config string
	parser := &parser{
		log:     log,
		homedir: homedir,
	}

	command := &cobra.Command{
		Use:   "start",
		Short: "Start all processes",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			settings, registry, err := parser.parseConfig(config)
			if err != nil {
				return err
			}

			if settings.LogFile != "" {
				f, err := os.OpenFile(settings.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
				if err != nil {
					return err
				}
				defer f.Close()

				log.w = f
			}

			runner := &processor{
				log:       log,
				reg:       registry,
				terminate: make(chan []string, len(registry.processes())),
				errors:    make(chan error, len(registry.processes())),
			}
			defer runner.shutdown()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

			go func() {
				runner.perform(ctx)
				stop()
			}()

			<-ctx.Done()
			return nil
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
	log     *logger
	homedir string
}

type settings struct {
	LogFile string `yaml:"log_file"`
}

func (ps *parser) parseConfig(path string) (*settings, *registry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}

	raw := make(map[string]any)
	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, nil, err
	}

	settings, err := ps.parseSettings(raw["settings"])
	if err != nil {
		return nil, nil, err
	}

	reg := newRegistry()
	services, err := ps.parseServices(raw["services"])
	if err != nil {
		return nil, nil, err
	}

	for name, p := range services {
		p.Name = name
		p.Done = make(chan struct{})
		p.Logger = ps.log
		if len(p.Hooks["wait"]) != 0 {
			p.waiting, p.doneWaiting = context.WithCancel(context.Background())
		}
		p.homedir = ps.homedir
		reg.register(p)
	}

	return settings, reg, nil
}

func (ps *parser) parseSettings(value any) (*settings, error) {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize settings")
	}

	// defaults
	cfg := settings{}

	err = yaml.Unmarshal(raw, &cfg)
	return &cfg, errors.Wrap(err, "could not parse settings")
}

func (ps *parser) parseServices(value any) (map[string]*process, error) {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, "could not serialize services")
	}

	services := make(map[string]*process)
	err = yaml.Unmarshal(raw, &services)
	return services, errors.Wrap(err, "could not parse services")
}
