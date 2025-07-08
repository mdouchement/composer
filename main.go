package main

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	var (
		verbose bool
		log     = &logger{w: os.Stdout}
		homedir string
	)

	//
	//

	c := &cobra.Command{
		Use:     "composer",
		Short:   "An awesome utility to manage all your processes in development environment",
		Version: Version(),
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) (err error) {
			homedir, err = os.Getwd()
			if err != nil {
				return err
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
			log.WithPrefixName("version").Info(c.Version)
		},
	})

	//
	//

	if err := c.Execute(); err != nil {
		log.WithPrefixName("composer").Error(err)
		time.Sleep(100 * time.Millisecond) // Wait logger oututing
		os.Exit(1)
	}
}
