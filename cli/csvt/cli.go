package csvt

import (
	"fmt"
	"path/filepath"

	"github.com/iancoleman/strcase"

	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/taskie/csvt"
	"github.com/taskie/jc"
	"github.com/taskie/osplus"
)

type Config struct {
	Mode, FromType, ToType, FromDelimiter, ToDelimiter, LogLevel string
	Error                                                        bool
}

var configFile string
var config Config
var (
	verbose, debug, version bool
)

const CommandName = "csvt"

func init() {
	Command.PersistentFlags().StringVarP(&configFile, "config", "c", "", `config file (default "`+CommandName+`.yml")`)
	Command.Flags().StringP("from-type", "f", "", "from type")
	Command.Flags().StringP("to-type", "t", "", "to type")
	Command.Flags().StringP("from-delimiter", "d", ",", "from delimiter")
	Command.Flags().StringP("to-delimiter", "D", ",", "to delimiter")
	Command.Flags().StringP("mode", "m", "", "mode")
	Command.Flags().BoolP("error", "e", false, "exit if error")
	Command.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	Command.Flags().BoolVar(&debug, "debug", false, "debug output")
	Command.Flags().BoolVarP(&version, "version", "V", false, "show Version")

	for _, s := range []string{"from-type", "to-type", "from-delimiter", "to-delimiter", "mode", "error"} {
		envKey := strcase.ToSnake(s)
		structKey := strcase.ToCamel(s)
		viper.BindPFlag(envKey, Command.Flags().Lookup(s))
		viper.RegisterAlias(structKey, envKey)
	}

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if debug {
		log.SetLevel(log.DebugLevel)
	} else if verbose {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName(CommandName)
		conf, err := osplus.GetXdgConfigHome()
		if err != nil {
			log.Info(err)
		} else {
			viper.AddConfigPath(filepath.Join(conf, CommandName))
		}
		viper.AddConfigPath(".")
	}
	viper.SetEnvPrefix(CommandName)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Debug(err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Warn(err)
	}
}

func Main() {
	Command.Execute()
}

var Command = &cobra.Command{
	Use:  CommandName + ` [INPUT] [OUTPUT]`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		err := run(cmd, args)
		if err != nil {
			log.Fatal(err)
		}
	},
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

func run(cmd *cobra.Command, args []string) error {
	if version {
		fmt.Println(csvt.Version)
		return nil
	}
	if config.LogLevel != "" {
		lv, err := log.ParseLevel(config.LogLevel)
		if err != nil {
			log.Warn(err)
		} else {
			log.SetLevel(lv)
		}
	}
	if debug {
		if viper.ConfigFileUsed() != "" {
			log.Debugf("Using config file: %s", viper.ConfigFileUsed())
		}
		log.Debug(pp.Sprint(config))
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
		return fmt.Errorf("invalid arguments: %v", args[2:])
	}

	fromType := config.FromType
	if fromType == "" {
		fromType = jc.ExtToType(filepath.Ext(input))
		if fromType == "" {
			if config.Mode == "unmap" {
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
			if config.Mode == "map" {
				toType = "json"
			} else {
				toType = "csv"
			}
		}
	}

	opener := osplus.NewOpener()
	r, err := opener.Open(input)
	if err != nil {
		return err
	}
	defer r.Close()
	w, commit, err := opener.CreateTempFileWithDestination(output, "", CommandName+"-")
	if err != nil {
		return err
	}
	defer w.Close()

	app := csvt.NewApplication(config.Mode)
	app.FromType = fromType
	app.ToType = toType
	app.FromDelimiter, err = firstRune(config.FromDelimiter)
	if err != nil {
		return err
	}
	app.ToDelimiter, err = firstRune(config.ToDelimiter)
	if err != nil {
		return err
	}
	err = app.Run(r, w)
	if err != nil {
		return err
	}

	commit(true)
	return nil
}
