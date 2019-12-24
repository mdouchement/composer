package main

import (
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

func main() {
	var (
		verbose bool
		log     = logrus.New()
		homedir string
	)

	//
	//

	c := &cobra.Command{
		Use:     "composer",
		Short:   "An awesome utility to manage all your processes in development environment",
		Version: fmt.Sprintf("%s - build %.7s @ %s", version, revision, date),
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) (err error) {
			homedir, err = os.Getwd()
			if err != nil {
				return err
			}

			log.SetOutput(os.Stdout)
			log.Formatter = new(prefixer.TextFormatter)

			if verbose || os.Getenv("APP_DEBUG") == "1" {
				log.Level = logrus.DebugLevel
			}

			return nil
		},
	}
	c.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Increase logger level")

	c.AddCommand(command(log, homedir))
	c.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Version for composer",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(c.Version)
		},
	})

	//
	//

	if err := c.Execute(); err != nil {
		log.Error(err)
		time.Sleep(100 * time.Millisecond) // Wait logger outputing
		os.Exit(1)
	}
}
