package main

import (
	"bufio"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

type (
	configuration struct {
		Logger logConfiguration `yaml:"logger"`
	}

	logConfiguration struct {
		BufferSize   int `yaml:"buffer_size"`
		EntryMaxSize int `yaml:"entry_max_size"`
	}
)

var (
	log     = logrus.New()
	homedir string
	cfg     configuration
)

func init() {
	var err error
	homedir, err = os.Getwd()
	check(err)

	cfg = configuration{
		Logger: logConfiguration{
			BufferSize:   42,
			EntryMaxSize: bufio.MaxScanTokenSize,
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Composer - MIT"
	app.Version = "0.2.0"
	app.Author = "mdouchement"
	app.Usage = "Usage:"
	app.Flags = globalFlags
	app.Before = beforeAction

	commands(app)

	err := app.Run(os.Args)
	check(err)
}

var globalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "D, verbose",
		Usage: "Increase logger level",
	},
}

func beforeAction(context *cli.Context) error {
	if context.Bool("D") || os.Getenv("APP_DEBUG") == "1" {
		log.Level = logrus.DebugLevel
	}
	return nil
}

func check(err error) {
	if err != nil {
		fail(err.Error())
	}
}

func fail(err string) {
	log.Error(err)
	time.Sleep(100 * time.Millisecond) // Wait logger outputing
	os.Exit(1)
}
