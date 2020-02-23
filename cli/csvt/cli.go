package csvt

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/taskie/csvt"
	"github.com/taskie/fwv"
	"github.com/taskie/jc"
	"github.com/taskie/ose"
	"github.com/taskie/ose/coli"
	"go.uber.org/zap"
)

const CommandName = "csvt"

var Command *cobra.Command

func init() {
	Command = NewCommand(coli.NewColiInThisWorld())
}

func Main() {
	Command.Execute()
}

func NewCommand(cl *coli.Coli) *cobra.Command {
	cmd := &cobra.Command{
		Use:  CommandName,
		Args: cobra.RangeArgs(0, 2),
		Run:  cl.WrapRun(run),
	}
	cl.Prepare(cmd)

	flg := cmd.Flags()
	flg.StringP("from-type", "f", "", "from type")
	flg.StringP("to-type", "t", "", "to type")
	flg.StringP("from-delimiter", "d", ",", "from delimiter")
	flg.StringP("to-delimiter", "D", ",", "to delimiter")
	flg.StringP("mode", "m", "", "mode [convert|transpose|map|unmap]")
	flg.BoolP("error", "e", false, "exit if error")

	cl.BindFlags(flg, []string{"from-type", "to-type", "from-delimiter", "to-delimiter", "mode", "error"})
	return cmd
}

type Config struct {
	Mode, FromType, ToType, FromDelimiter, ToDelimiter, LogLevel string
	Error                                                        bool
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

func run(cl *coli.Coli, cmd *cobra.Command, args []string) {
	v := cl.Viper()
	log := zap.L()
	if v.GetBool("version") {
		cmd.Println(fwv.Version)
		return
	}
	var config Config
	err := v.Unmarshal(&config)
	if err != nil {
		log.Fatal("can't unmarshal config", zap.Error(err))
	}

	input := ""
	output := ""
	switch len(args) {
	case 0:
		break
	case 1:
		input = args[0]
	case 2:
		input = args[0]
		output = args[1]
	default:
		log.Fatal("invalid arguments", zap.Strings("arguments", args[2:]))
	}

	mode := config.Mode
	if mode == "" {
		mode = "convert"
	}

	fromType := config.FromType
	if fromType == "" {
		fromType = jc.ExtToType(filepath.Ext(input))
		if fromType == "" {
			if mode == "unmap" {
				fromType = "json"
			} else {
				fromType = "csv"
			}
		}
	}
	toType := config.ToType
	if toType == "" {
		toType = jc.ExtToType(filepath.Ext(output))
		if toType == "" {
			if mode == "map" {
				toType = "json"
			} else {
				toType = "csv"
			}
		}
	}

	opener := ose.NewOpenerInThisWorld()
	r, err := opener.Open(input)
	if err != nil {
		log.Fatal("can't open", zap.Error(err))
	}
	defer r.Close()
	_, err = opener.CreateTempFile("", CommandName, output, func(f io.WriteCloser) (bool, error) {
		app := csvt.NewApplication(mode)
		app.FromType = fromType
		app.ToType = toType
		app.FromDelimiter, err = firstRune(config.FromDelimiter)
		if err != nil {
			return false, err
		}
		app.ToDelimiter, err = firstRune(config.ToDelimiter)
		if err != nil {
			return false, err
		}
		err = app.Run(r, f)
		if err != nil {
			log.Fatal("can't convert", zap.Error(err))
		}
		return true, nil
	})
	if err != nil {
		log.Fatal("can't create file", zap.Error(err))
	}
}
