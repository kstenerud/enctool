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

func (this *cmdPrint) Name() string { return "print" }

func (this *cmdPrint) Description() string { return "Print a document's structure" }

func (this *cmdPrint) Usage() string {
	fs, _ := this.newFlagSet()
	return getFlagsUsage(fs)
}

func (this *cmdPrint) Run() (err error) {
	v, err := this.decode(this.srcReader)
	if err != nil {
		return
	}
	if v == nil {
		fmt.Println("<nil>")
	} else {
		fmt.Println(describe.Describe(v, int(this.indent)))
	}
	return
}

func (this *cmdPrint) Init(args []string) (err error) {
	fs, fields := this.newFlagSet()
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

	this.srcReader, err = openFileRead(srcFile)
	if err != nil {
		return err
	}

	if len(srcFormat) == 0 {
		bufReader := bufio.NewReader(this.srcReader)
		this.srcReader = bufReader
		srcFormat, err = detectSrcFormat(bufReader)
		if err != nil {
			return err
		}
	}

	this.indent = fields.getUint("i")

	this.decode, err = getDecoder(srcFormat)
	if err != nil {
		return err
	}

	return
}

func (this *cmdPrint) newFlagSet() (fs *flag.FlagSet, fields fieldValues) {
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
