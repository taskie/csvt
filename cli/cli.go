package cli

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/taskie/csvt"
	"github.com/taskie/jc"
	"github.com/taskie/osplus"
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

func pathToType(fpath string) string {
	return jc.ExtToType(filepath.Ext(fpath))
}

func Main() {
	var opts Options
	args, err := flags.ParseArgs(&opts, os.Args)
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
	inputPath := ""
	inputExtType := ""
	if len(args) > 1 {
		inputPath = args[1]
		inputExtType = pathToType(inputPath)
	}
	outputPath := ""
	outputExtType := ""
	if len(args) > 2 {
		outputPath = args[2]
		outputExtType = pathToType(outputPath)
	}
	app := csvt.NewApplication(opts.Mode)
	app.FromDelimiter = fromDelimiter
	app.ToDelimiter = toDelimiter
	if opts.FromType != "" {
		app.FromType = opts.FromType
	} else if inputExtType != "" {
		app.FromType = inputExtType
	} else if opts.Mode == "unmap" {
		app.FromType = "json"
	}
	if opts.ToType != "" {
		app.ToType = opts.ToType
	} else if outputExtType != "" {
		app.ToType = outputExtType
	} else if opts.Mode == "map" {
		app.ToType = "json"
	}
	var r io.Reader
	if inputPath == "" || inputPath == "-" {
		r = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(inputPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer file.Close()
		r = bufio.NewReader(file)
	}
	if outputPath == "" || outputPath == "-" {
		err = app.Run(r, os.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		tmpFile, err := ioutil.TempFile("", "csvt-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer os.Remove(tmpFile.Name())
		err = func() error {
			defer tmpFile.Close()
			w := bufio.NewWriter(tmpFile)
			defer w.Flush()
			return app.Run(r, w)
		}()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = osplus.Copy(tmpFile.Name(), outputPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
