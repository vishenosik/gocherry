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

func AppFlags(writer io.Writer) func(*flagset) {
	return func(f *flagset) {
		f.BoolFunc("version", "Show build info", func(s string) error {
			BuildInfoYaml(writer)
			return ErrSuccessExit
		})
	}
}

func ConfigFlags(writer io.Writer, structs ...any) func(*flagset) {

	return func(f *flagset) {
		f.BoolFunc("config.info", "Show config schema information", FlagConfigInfoEnv(writer, structs...))
		f.Func("config.gen", "Generate config schema", FlagConfigGenEnv(structs...))
	}
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

func Flags(writer io.Writer, args []string, flagsets ...func(*flagset)) {
	if err := parseFlags(writer, args, flagsets...); err != nil {
		if errors.Is(err, flag.ErrHelp) || errors.Is(err, ErrSuccessExit) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func parseFlags(writer io.Writer, args []string, flagsets ...func(*flagset)) error {

	flags := flag.NewFlagSet("app flags", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	flagSet := &flagset{
		writer:  writer,
		FlagSet: flags,
	}

	for _, flagset := range flagsets {
		flagset(flagSet)
	}

	return flagSet.Parse(args)
}
