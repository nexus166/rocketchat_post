package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/szampardi/hermes"
)

var (
	showVersion           bool
	semver, commit, built = "v0.0.0-dev", "local", "a while ago" //
)

func init() {
	hermes.CLIFlags.BoolVar(&showVersion, "v", false, "print build version/date and exit")
	for !hermes.CLIFlags.Parsed() {
		if err := hermes.CLIFlags.Parse(os.Args[1:]); err != nil && err != flag.ErrHelp {
			panic(err)
		}
	}
	if showVersion {
		fmt.Fprintf(os.Stderr, "github.com/szampardi/hermes version %s (%s) built %s\n", semver, commit, built)
		os.Exit(0)
	}
}

func main() {
	var data hermes.Data
	if stdin, err := os.Stdin.Stat(); err == nil && (stdin.Mode()&os.ModeCharDevice) == 0 {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %s\n", stdin.Name(), err)
		} else {
			data.Stdin = string(b)
		}
	}
	data.Args = hermes.CLIFlags.Args()
	if data.Stdin == "" && len(data.Args) < 1 {
		panic("nothing to send")
	}
	if err := hermes.Post(data); err != nil {
		panic(err)
	}
}
