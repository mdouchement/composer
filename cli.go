package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	command.Flags().StringVarP(&config, "config", "c", "", "Configuration file")
}

var (
	command = &cobra.Command{
		Use:   "start",
		Short: "Start all processes",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) (err error) {
			registry, err := parseConfig(config)
			if err != nil {
				return err
			}

			//

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)
			go func() {
				<-signals
				log.Info("Gracefully shutdown composer")

				processes := registry.processes()
				names := make([]string, len(processes))
				for i, process := range processes {
					names[i] = process.Name
				}

				stopAllGivenNames(registry, names)
			}()

			//

			perform(registry)
			return nil
		},
	}

	config string
)

func parseConfig(path string) (*registry, error) {
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

	parseSettings(raw["settings"])

	reg := newRegistry()
	for name, p := range parseServices(raw["services"]) {
		p.Name = name
		p.Done = make(chan struct{})
		p.Logger = logrus.NewEntry(log)
		reg.register(p)
	}

	return reg, nil
}

func parseServices(value interface{}) map[string]*process {
	services := make(map[string]*process)

	raw, err := yaml.Marshal(value)
	if err != nil {
		fail(fmt.Sprintf("services marshalling: %s", err))
	}

	err = yaml.Unmarshal(raw, &services)
	if err != nil {
		fail(fmt.Sprintf("services unmarshalling: %s", err))
	}

	return services
}

func parseSettings(value interface{}) {
	raw, err := yaml.Marshal(value)
	if err != nil {
		fail(fmt.Sprintf("settings marshalling: %s", err))
	}

	err = yaml.Unmarshal(raw, &cfg)
	if err != nil {
		fail(fmt.Sprintf("settings unmarshalling: %s", err))
	}
}
