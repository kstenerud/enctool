// Copyright 2020 Karl Stenerud
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
)

func getFlagsUsage(fs *flag.FlagSet) string {
	// Workaround for poor design of FlagSet: Redirect automatic output.
	outputBuffer := bytes.Buffer{}
	fs.SetOutput(&outputBuffer)
	fs.Usage()
	return string(outputBuffer.Bytes())
}

func parseFlagsQuietly(fs *flag.FlagSet, args []string) (err error) {
	// Workaround for poor design of FlagSet: Stop automatic output.
	junkBuffer := bytes.Buffer{}
	fs.SetOutput(&junkBuffer)
	err = fs.Parse(args)
	return
}

func usageError(format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%v%w", message, UsageError)
}

func openFileRead(path string) (io.Reader, error) {
	if path == "-" {
		return os.Stdin, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func openFileWrite(path string) (io.Writer, error) {
	if path == "-" {
		return os.Stdout, nil
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type fieldValues map[string]interface{}

func (this fieldValues) getRequiredString(id string, name string) (string, error) {
	field, ok := this[id]
	if !ok {
		return "", usageError("%v (-%v) is a required parameter", name, id)
	}
	v, ok := field.(*string)
	if !ok {
		panic(fmt.Errorf("BUG: Expected %v to contain a string, but contains %v", id, field))
	}
	value := *v
	if value == "" {
		return "", usageError("%v (-%v) is a required parameter", name, id)
	}
	return value, nil
}

func (this fieldValues) getString(id string, name string) (string, error) {
	field, ok := this[id]
	if !ok {
		return "", nil
	}
	v, ok := field.(*string)
	if !ok {
		panic(fmt.Errorf("BUG: Expected %v to contain a string, but contains %v", id, field))
	}
	return *v, nil
}

func (this fieldValues) getUint(id string) uint {
	field := this[id]
	v, ok := field.(*uint)
	if !ok {
		panic(fmt.Errorf("BUG: Expected %v to contain a uint, but contains %v (%v)", id, field, reflect.TypeOf(field)))
	}
	return *v
}

func detectSrcFormat(reader *bufio.Reader) (string, error) {
	begin, err := reader.Peek(1)
	if err != nil {
		return "", err
	}
	switch begin[0] {
	case 'c':
		return "cte", nil
	case 0x03:
		return "cbe", nil
	default:
		return "json", nil
	}
}
