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
	"flag"
	"fmt"
	"io"

	"github.com/kstenerud/go-describe"
)

type cmdPrint struct {
	srcReader io.Reader
	decode    decoder
	indent    uint
}

func (_this *cmdPrint) Name() string { return "print" }

func (_this *cmdPrint) Description() string { return "Print a document's structure" }

func (_this *cmdPrint) Usage() string {
	fs, _ := _this.newFlagSet()
	return getFlagsUsage(fs)
}

func (_this *cmdPrint) Run() (err error) {
	v, err := _this.decode(_this.srcReader)
	if err != nil {
		return
	}
	if v == nil {
		fmt.Println("<nil>")
	} else {
		fmt.Println(describe.Describe(v, int(_this.indent)))
	}
	return
}

func (_this *cmdPrint) Init(args []string) (err error) {
	fs, fields := _this.newFlagSet()
	if err = parseFlagsQuietly(fs, args); err != nil {
		return usageError("%v", err)
	}

	srcFile, err := fields.getString("f", "File")
	if err != nil {
		return
	}
	if srcFile == "" {
		srcFile = "-"
	}

	srcFormat, err := fields.getString("fmt", "Format")
	if err != nil {
		return
	}

	_this.srcReader, err = openFileRead(srcFile)
	if err != nil {
		return err
	}

	if len(srcFormat) == 0 {
		bufReader := bufio.NewReader(_this.srcReader)
		_this.srcReader = bufReader
		srcFormat, err = detectSrcFormat(bufReader)
		if err != nil {
			return err
		}
	}

	_this.indent = fields.getUint("i")

	_this.decode, err = getDecoder(srcFormat)
	if err != nil {
		return err
	}

	return
}

func (_this *cmdPrint) newFlagSet() (fs *flag.FlagSet, fields fieldValues) {
	fields = make(fieldValues)
	fs = flag.NewFlagSet("print", flag.ContinueOnError)
	fields["fmt"] = fs.String("fmt", "", "File format (auto-detected if not specified)")
	fields["f"] = fs.String("f", "", "File to read from (- for stdin) (defaults to stdin)")
	fields["i"] = fs.Uint("i", 0, "Indentation (spaces)")

	return
}

func init() {
	addCommand(new(cmdPrint))
}
