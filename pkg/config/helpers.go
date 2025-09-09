package config

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type StructsManager struct {
	structs []any
}

var _structs *StructsManager

func Manager() *StructsManager {
	if _structs == nil {
		_structs = &StructsManager{
			structs: make([]any, 0),
		}
	}
	return _structs
}

func AddStructs(structs ...any) {
	Manager().structs = append(Manager().structs, structs...)
}

func (sm *StructsManager) Cleanup() {
	sm.structs = nil
}

func Structs() []any {
	return _structs.structs
}

const (
	whiteSpace   = 32
	headerFormat = "\n#=== %s ===#\n\n"
)

type Header interface {
	Desc() string
}

func ConfigInfoEnv(writer io.Writer, structs ...any) {

	structs = append(structs, Structs()...)

	writeHeader := func(header string) {
		_, _ = writer.Write(fmt.Appendf([]byte{}, headerFormat, header))
	}

	for _, _struct := range structs {
		if header, ok := _struct.(Header); ok {
			writeHeader(header.Desc())
		} else {
			writeHeader(reflect.TypeOf(_struct).String())
		}
		writer.Write(genEnvConfig(_struct))
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

		descTag, ok := field.Tag.Lookup("desc")
		if !ok || descTag == "-" {
			continue
		}

		writer.Write(fmt.Appendf([]byte{}, "# %s (%s)\n", descTag, field.Type))

		if envTag, ok := field.Tag.Lookup("env"); ok {
			writer.Write([]byte(envTag + "="))
		}

		if defaultTag, ok := field.Tag.Lookup("env-default"); ok {
			writer.Write([]byte(defaultTag))
		}

		writer.Write([]byte("\n"))

	}

}
