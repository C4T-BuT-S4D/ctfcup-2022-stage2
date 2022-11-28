package main

import (
	"os"

	"gmtchecker/checker"

	"github.com/pomo-mondreganto/go-checklib"
)

func main() {
	os.Exit(checklib.Run(&checker.Checker{}))
}
