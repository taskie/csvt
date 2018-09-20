package csvt

import (
	"encoding/csv"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/taskie/jc"
	"io"
	"os"
)

var (
	version  string
	revision string
)

type Csvt struct {
	FromType      string
	ToType        string
	FromDelimiter rune
	ToDelimiter   rune
	TransformMode string
}

func (csvt *Csvt) ReadAll(r io.Reader) (interface{}, error) {
	switch csvt.FromType {
	case "csv":
		csvr := csv.NewReader(r)
		csvr.Comma = csvt.FromDelimiter
		records, err := csvr.ReadAll()
		return records, err
	default:
		var data interface{}
		jcr := jc.Jc{
			FromType: csvt.FromType,
		}
		err := jcr.Decode(r, &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
}

func (csvt *Csvt) WriteAll(w io.Writer, records [][]string) error {
	switch csvt.ToType {
	case "csv":
		csvw := csv.NewWriter(w)
		csvw.Comma = csvt.ToDelimiter
		return csvw.WriteAll(records)
	default:
		if csvt.TransformMode == "titled" {
			titledItems := make([]map[string]interface{}, 0)
			if len(records) >= 1 {
				titleRecord := records[0]
				for i := 1; i < len(records); i++ {
					titledItem := make(map[string]interface{})
					record := records[i]
					for j, cell := range record {
						titledItem[titleRecord[j]] = cell
					}
					titledItems = append(titledItems, titledItem)
				}
			}
			jcw := jc.Jc{
				ToType: csvt.ToType,
			}
			return jcw.Encode(w, titledItems)
		} else {
			jcw := jc.Jc{
				ToType: csvt.ToType,
			}
			return jcw.Encode(w, records)
		}
	}
}

func (csvt *Csvt) Run(r io.Reader, w io.Writer) error {
	var err error
	data, err := csvt.ReadAll(r)
	if err != nil {
		return err
	}
	var records = make([][]string, 0, 0)
	if v, ok := data.([][]string); ok {
		records = v
	} else {
		// TODO
	}
	err = csvt.WriteAll(w, records)
	return err
}

type Options struct {
	TransformMode string `short:"m" long:"mode" default:"auto" description:"transform mode"`
	FromType      string `short:"f" long:"from" default:"csv" description:"convert from [csv|json|toml|msgpack]"`
	ToType        string `short:"t" long:"to" default:"json" description:"convert to [csv|json|toml|yaml|msgpack]"`
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

func Main(args []string) {
	var opts Options
	args, err := flags.ParseArgs(&opts, args)
	if opts.Version {
		if opts.Verbose {
			fmt.Println("Version: ", version)
			fmt.Println("Revision: ", revision)
		} else {
			fmt.Println(version)
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
	csvt := Csvt{
		FromType:      opts.FromType,
		ToType:        opts.ToType,
		FromDelimiter: fromDelimiter,
		ToDelimiter:   toDelimiter,
		TransformMode: opts.TransformMode,
	}
	err = csvt.Run(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
