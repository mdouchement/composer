package main

import (
	"fmt"
	"runtime"
)

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"
)

func Version() string {
	return fmt.Sprintf("%s (revision %.7s @ %s) - %s", version, revision, date, runtime.Version())
}
