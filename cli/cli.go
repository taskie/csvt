package cli

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/taskie/csvt"
	"os"
)

type Options struct {
	Mode          string `short:"m" long:"mode" default:"convert" description:"transform mode [convert|transpose|map|unmap]"`
	FromType      string `short:"f" long:"from" description:"convert from [csv|json|yaml|msgpack]"`
	ToType        string `short:"t" long:"to" description:"convert to [csv|json|yaml|msgpack]"`
	FromDelimiter string `short:"d" long:"from-delimiter" default:"," description:"delimiter (--from csv)"`
	ToDelimiter   string `short:"D" long:"to-delimiter" default:"," description:"delimiter (--to csv)"`
	NoColor       bool   `long:"no-color" env:"NO_COLOR" description:"NOT colorize output"`
	Verbose       bool   `short:"v" long:"verbose" description:"show verbose output"`
	Version       bool   `short:"V" long:"version" description:"show version"`
}

func firstRune(s string) (rune, error) {
	if len(s) != 1 {
		return '\x00', fmt.Errorf("delimiter must be a single character: %s", s)
	}
	for _, c := range s {
		return c, nil
	}
	panic("unreachable")
}

func Main() {
	var opts Options
	_, err := flags.ParseArgs(&opts, os.Args)
	if opts.Version {
		if opts.Verbose {
			fmt.Println("Version: ", csvt.Version)
			if csvt.Revision != "" {
				fmt.Println("Revision: ", csvt.Revision)
			}
		} else {
			fmt.Println(csvt.Version)
		}
		os.Exit(0)
	}
	fromDelimiter, err := firstRune(opts.FromDelimiter)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	toDelimiter, err := firstRune(opts.ToDelimiter)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app := csvt.NewApplication(opts.Mode)
	app.FromDelimiter = fromDelimiter
	app.ToDelimiter = toDelimiter
	if opts.FromType != "" {
		app.FromType = opts.FromType
	} else if opts.Mode == "unmap" {
		app.FromType = "json"
	}
	if opts.ToType != "" {
		app.ToType = opts.ToType
	} else if opts.Mode == "map" {
		app.ToType = "json"
	}
	err = app.Run(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
