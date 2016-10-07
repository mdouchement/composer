package main

import (
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
	perform(registry)

	return nil
}

func parseConfig(path string) *registry {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	check(err)

	rps := make(map[string]*process)
	err = yaml.Unmarshal(data, &rps)
	check(err)

	reg := newRegistry()
	go reg.statusToProfiler()
	for name, p := range rps {
		p.Name = name
		p.Done = make(chan struct{})
		reg.register(p)
	}

	return reg
}
