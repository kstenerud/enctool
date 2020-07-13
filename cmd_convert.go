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
	"strings"
)

type cmdConvert struct {
	srcReader     io.Reader
	dstWriter     io.Writer
	converter     converter
	encode        encoder
	decode        decoder
	encoderConfig encoderConfig
}

func (this *cmdConvert) Name() string { return "convert" }

func (this *cmdConvert) Description() string { return "Convert between formats" }

func (this *cmdConvert) Usage() string {
	fs, _ := this.newFlagSet()
	return getFlagsUsage(fs)
}

func (this *cmdConvert) Run() (err error) {
	err = this.converter(this.srcReader, this.dstWriter, &this.encoderConfig)
	return
}

func (this *cmdConvert) Init(args []string) (err error) {
	fs, fields := this.newFlagSet()
	if err = parseFlagsQuietly(fs, args); err != nil {
		return usageError("%v", err)
	}

	srcFile, err := fields.getRequiredString("s", "Source file")
	if err != nil {
		return
	}
	dstFile, err := fields.getRequiredString("d", "Destination file")
	if err != nil {
		return
	}
	srcFormat, err := fields.getRequiredString("sf", "Source format")
	if err != nil {
		return
	}
	dstFormat, err := fields.getRequiredString("df", "Destination format")
	if err != nil {
		return
	}

	this.encoderConfig.indentSpaces = int(fields.getUint("i"))

	converterID := strings.ToLower(srcFormat) + "-" + strings.ToLower(dstFormat)

	this.converter, err = getConverter(converterID)
	if err != nil {
		return err
	}

	this.srcReader, err = openFileRead(srcFile)
	if err != nil {
		return err
	}

	this.dstWriter, err = openFileWrite(dstFile)

	return
}

func (this *cmdConvert) newFlagSet() (fs *flag.FlagSet, fields fieldValues) {
	fields = make(fieldValues)
	fs = flag.NewFlagSet("convert", flag.ContinueOnError)
	fields["sf"] = fs.String("sf", "", "The source format to convert from (- for stdin) (required)")
	fields["df"] = fs.String("df", "", "The destination format to convert to (- for stdout) (required)")
	fields["s"] = fs.String("s", "", "The source file to read from (required)")
	fields["d"] = fs.String("d", "", "The destination file to write to (required)")
	fields["i"] = fs.Uint("i", 0, "Indentation to use")

	return
}

func init() {
	addCommand(new(cmdConvert))
}
