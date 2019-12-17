package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixer "github.com/x-cray/logrus-prefixed-formatter"
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
	version  = "dev"
	revision = "none"
	date     = "unknown"
)

var (
	verbose bool
	log     = logrus.New()
	homedir string
	cfg     configuration
)

func init() {
	var err error
	homedir, err = os.Getwd()
	check(err)

	log.Formatter = new(prefixer.TextFormatter)

	cfg = configuration{
		Logger: logConfiguration{
			BufferSize:   42,
			EntryMaxSize: bufio.MaxScanTokenSize,
		},
	}
}

func main() {
	c := &cobra.Command{
		Use:     "composer",
		Short:   "An awesome utility to manage all your processes in development environment",
		Version: fmt.Sprintf("%s - build %.7s @ %s", version, revision, date),
		Args:    cobra.NoArgs,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if verbose || os.Getenv("APP_DEBUG") == "1" {
				log.Level = logrus.DebugLevel
			}
		},
	}
	c.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Increase logger level")

	c.AddCommand(command)
	c.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Version for composer",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(c.Version)
		},
	})

	if err := c.Execute(); err != nil {
		log.Error(err)
		time.Sleep(100 * time.Millisecond) // Wait logger outputing
		os.Exit(1)
	}
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
