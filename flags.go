package gocherry

import (
	"errors"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/vishenosik/gocherry/pkg/config"
)

var (
	ErrSuccessExit = errors.New("success and exit with code 0")
)

type flagset struct {
	writer io.Writer
	*flag.FlagSet
}

func NewFlagSet(name string, errorHandling flag.ErrorHandling, writer io.Writer) *flagset {
	flags := flag.NewFlagSet(name, errorHandling)
	flags.SetOutput(io.Discard)
	return &flagset{
		writer:  writer,
		FlagSet: flags,
	}
}

func (f *flagset) Parse(arguments []string) error {

	print := func(msg string) {
		f.SetOutput(f.writer)
		f.writer.Write([]byte(msg + "\nUsage of config flags\n"))
		f.PrintDefaults()
	}

	if err := f.FlagSet.Parse(arguments); err != nil {
		// flag.Parse doesn't return wrapped error, so errors.Is() assertion doesn't work here
		if strings.Contains(err.Error(), ErrSuccessExit.Error()) {
			return ErrSuccessExit
		}

		print(err.Error())
		return err
	}

	return nil
}

func AppFlags(writer io.Writer, args []string) {

	flags := NewFlagSet("app flags", flag.ContinueOnError, writer)

	flags.BoolFunc("version", "Show build info", func(s string) error {
		BuildInfoYaml(writer)
		return ErrSuccessExit
	})

	flags.Parse(args)
}

func ConfigFlags(writer io.Writer, args []string, structs ...any) {
	var exitCode int

	defer func() {
		os.Exit(exitCode)
	}()

	if err := configFlags(writer, args, structs...); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			exitCode = 2
			return
		}

		if errors.Is(err, ErrSuccessExit) {
			exitCode = 2
			return
		}

		exitCode = 1
	}
}

func configFlags(writer io.Writer, args []string, structs ...any) error {

	flags := NewFlagSet("config flags", flag.ContinueOnError, writer)
	flags.SetOutput(io.Discard)

	_structs := append(structs, config.Structs()...)
	flags.BoolFunc("config.info", "Show config schema information", FlagConfigInfoEnv(writer, _structs...))
	flags.Func("config.gen", "Generate config schema", FlagConfigGenEnv(_structs...))

	return flags.Parse(args)
}

func FlagConfigInfoEnv(writer io.Writer, structs ...any) func(string) error {
	return func(string) error {
		config.ConfigInfoEnv(writer, structs...)
		return ErrSuccessExit
	}
}

func FlagConfigGenEnv(structs ...any) func(string) error {
	return func(filename string) error {
		if filename == "" {
			filename = "example.env"
		}
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		return FlagConfigInfoEnv(file, structs...)(filename)
	}
}
