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

	srcFile, err := fields.getString("s", "Source file")
	if err != nil {
		return
	}
	if srcFile == "" {
		srcFile = "-"
	}

	dstFile, err := fields.getString("d", "Destination file")
	if err != nil {
		return
	}
	if dstFile == "" {
		dstFile = "-"
	}

	srcFormat, err := fields.getString("sf", "Source format")
	if err != nil {
		return
	}
	dstFormat, err := fields.getRequiredString("df", "Destination format")
	if err != nil {
		return
	}

	readAdapter := readAdapterNone
	if fields.getBool("x") {
		readAdapter = readAdapterHex
	}
	if fields.getBool("t") {
		if readAdapter != readAdapterNone {
			return fmt.Errorf("Cannot choose modes -x and -t simultaneously")
		}
		readAdapter = readAdapterText
	}

	writeAdapter := writeAdapterNone
	if fields.getBool("X") {
		writeAdapter = writeAdapterHex
	}
	if fields.getBool("C") {
		if writeAdapter != writeAdapterNone {
			return fmt.Errorf("Cannot choose more than one of -X -C -S simultaneously")
		}
		writeAdapter = writeAdapterC
	}
	if fields.getBool("S") {
		if writeAdapter != writeAdapterNone {
			return fmt.Errorf("Cannot choose more than one of -X -C -S simultaneously")
		}
		writeAdapter = writeAdapterStringify
	}

	this.encoderConfig.indentSpaces = int(fields.getUint("i"))

	this.srcReader, err = openFileRead(srcFile)
	if err != nil {
		return fmt.Errorf("Error opening %v: %v", srcFile, err)
	}
	switch readAdapter {
	case readAdapterHex:
		this.srcReader = newHexReader(this.srcReader)
	case readAdapterText:
		this.srcReader = newTextByteReader(this.srcReader)
	}

	if len(srcFormat) == 0 {
		bufReader := bufio.NewReader(this.srcReader)
		this.srcReader = bufReader
		srcFormat, err = detectSrcFormat(bufReader)
		if err != nil {
			return fmt.Errorf("Error detecting source format of %v: %v", srcFile, err)
		}
	}

	converterID := strings.ToLower(srcFormat) + "-" + strings.ToLower(dstFormat)

	this.converter, err = getConverter(converterID)
	if err != nil {
		return err
	}

	this.dstWriter, err = openFileWrite(dstFile)
	switch writeAdapter {
	case writeAdapterC:
		this.dstWriter = newCWriter(this.dstWriter)
	case writeAdapterHex:
		this.dstWriter = newHexWriter(this.dstWriter)
	case writeAdapterStringify:
		this.dstWriter = newStringifyWriter(this.dstWriter)
	}

	return
}

func (this *cmdConvert) newFlagSet() (fs *flag.FlagSet, fields fieldValues) {
	fields = make(fieldValues)
	fs = flag.NewFlagSet("convert", flag.ContinueOnError)
	fields["sf"] = fs.String("sf", "", "The source format to convert from (auto-detected if not specified)")
	fields["df"] = fs.String("df", "", "The destination format to convert to (required)")
	fields["s"] = fs.String("s", "", "The source file to read from (- for stdin) (defaults to stdin)")
	fields["d"] = fs.String("d", "", "The destination file to write to (- for stdout) (defaults to stdout)")
	fields["i"] = fs.Uint("i", 0, "Indentation to use")
	fields["t"] = fs.Bool("t", false, "Interpret source as text-encoded byte values (can be decimal numbers or 0xff style hex, separated by non-numeric chars)")
	fields["x"] = fs.Bool("x", false, "Interpret source as text hex-encoded byte values (2 digits per byte), separated by non-numeric chars")
	fields["X"] = fs.Bool("X", false, "write destination as text hex-encoded byte values (2 digits per byte), separated a space")
	fields["C"] = fs.Bool("C", false, "write destination as C-style byte values (in the format '0xab, 0xcd, ...')")
	fields["S"] = fs.Bool("S", false, "write destination as stringified (escapes \" and \\ characters)")

	return
}

func init() {
	addCommand(new(cmdConvert))
}

type readAdapter int

const (
	readAdapterNone = iota
	readAdapterHex
	readAdapterText
)

type writeAdapter int

const (
	writeAdapterNone = iota
	writeAdapterHex
	writeAdapterC
	writeAdapterStringify
)
