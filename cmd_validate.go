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
	"flag"
	"io"
)

type cmdValidate struct {
	srcReader io.Reader
	decode    decoder
}

func (this *cmdValidate) Name() string { return "validate" }

func (this *cmdValidate) Description() string { return "Validate a document." }

func (this *cmdValidate) Usage() string {
	fs, _ := this.newFlagSet()
	return getFlagsUsage(fs)
}

func (this *cmdValidate) Run() (err error) {
	_, err = this.decode(this.srcReader)
	return
}

func (this *cmdValidate) Init(args []string) (err error) {
	fs, fields := this.newFlagSet()
	if err = parseFlagsQuietly(fs, args); err != nil {
		return usageError("%v", err)
	}

	srcFile, err := fields.getRequiredString("f", "File (- for stdin)")
	if err != nil {
		return
	}
	srcFormat, err := fields.getRequiredString("fmt", "Format")
	if err != nil {
		return
	}

	this.decode, err = getDecoder(srcFormat)
	if err != nil {
		return err
	}

	this.srcReader, err = openFileRead(srcFile)
	if err != nil {
		return err
	}

	return
}

func (this *cmdValidate) newFlagSet() (fs *flag.FlagSet, fields fieldValues) {
	fields = make(fieldValues)
	fs = flag.NewFlagSet("validate", flag.ContinueOnError)
	fields["fmt"] = fs.String("fmt", "", "File format (required)")
	fields["f"] = fs.String("f", "", "File to read from (required)")

	return
}

func init() {
	addCommand(new(cmdValidate))
}
