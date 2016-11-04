package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func commands(app *cli.App) {
	app.Commands = []cli.Command{
		startCommand,
	}
}

var startCommand = cli.Command{
	Name:   "start",
	Usage:  "Start all processes",
	Action: startAction,
	Flags:  startFlags,
}

var startFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "c, config",
		Usage: "Path to the configuration file",
	},
}

func startAction(context *cli.Context) error {
	config := context.String("c")
	if config == "" {
		fail("Need configuration file: `-c` option")
	}

	registry := parseConfig(config)
	startLogger()
	perform(registry)

	return nil
}

func parseConfig(path string) *registry {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	check(err)

	raw := make(map[string]interface{})
	err = yaml.Unmarshal(data, &raw)
	check(err)

	parseSettings(raw["settings"])

	reg := newRegistry()
	for name, p := range parseServices(raw["services"]) {
		p.Name = name
		p.Done = make(chan struct{})
		reg.register(p)
	}

	return reg
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
