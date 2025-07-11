package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

var (
	_structs []any
)

func AddStructs(structs ...any) {
	if len(_structs) == 0 {
		_structs = make([]any, 0, len(structs))
	}
	_structs = append(_structs, structs...)
}

func Structs() []any {
	return _structs
}

const (
	whiteSpace = 32
)

type Header interface {
	Desc() string
}

func ConfigInfoEnv(writer io.Writer, structs ...any) func(string) error {
	return func(string) error {
		defer os.Exit(0)

		for _, _struct := range structs {

			if header, ok := _struct.(Header); ok {
				writeHeader(writer, header.Desc())
			} else {
				writeHeader(writer, reflect.TypeOf(_struct).String())
			}

			writer.Write(genEnvConfig(_struct))
		}
		return nil
	}
}

func writeHeader(writer io.Writer, header string) {
	writer.Write(fmt.Appendf([]byte{}, "\n#=== %s ===#\n\n", header))
}

func ConfigGenEnv(structs ...any) func(string) error {
	return func(filename string) error {
		if filename == "" {
			filename = "example.env"
		}
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		return ConfigInfoEnv(file, structs...)(filename)
	}
}

type StringerWriter interface {
	io.Writer
	fmt.Stringer
}

type indentWrapper struct {
	writer StringerWriter
	indent int
}

func newIndent(writer StringerWriter, indent int) *indentWrapper {
	return &indentWrapper{writer: writer, indent: indent}
}

func (i *indentWrapper) Write(p []byte) (n int, err error) {
	indenter := []byte{whiteSpace}
	cp := make([]byte, len(p))
	copy(cp, p)

	if bytes.Contains(cp, []byte("\n")) {
		return i.writer.Write(cp)
	}
	return i.writer.Write(append(cp, bytes.Repeat(indenter, i.indent)...))
}

func (i *indentWrapper) Bytes() []byte {
	return []byte(i.writer.String())
}

func genEnvConfig(cfg any) []byte {

	_type := reflect.TypeOf(cfg)

	if _type.Kind() == reflect.Interface {
		_type = _type.Elem()
	}

	if _type.Kind() == reflect.Pointer {
		_type = _type.Elem()
	}

	if _type.Kind() != reflect.Struct {
		return nil
	}

	builder := new(strings.Builder)

	writer := newIndent(builder, 0)
	genEnvConfigRecursively(builder, _type)

	return writer.Bytes()
}

func genEnvConfigRecursively(writer io.Writer, _type reflect.Type) {

	for i := range _type.NumField() {

		field := _type.Field(i)

		if field.Type.Kind() == reflect.Struct {
			genEnvConfigRecursively(writer, field.Type)
			continue
		}

		writer.Write([]byte("# "))
		descTag, ok := field.Tag.Lookup("desc")
		if ok {
			writer.Write([]byte(descTag))
		}
		writer.Write([]byte(fmt.Sprintf(" (%s)\n", field.Type)))

		if envTag, ok := field.Tag.Lookup("env"); ok {
			writer.Write([]byte(envTag + "="))
		}

		if defaultTag, ok := field.Tag.Lookup("default"); ok {
			writer.Write([]byte(defaultTag))
		}

		writer.Write([]byte("\n"))

	}

}
