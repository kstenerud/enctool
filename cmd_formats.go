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
	"fmt"
)

type cmdFormats struct {
}

func (this *cmdFormats) Name() string { return "formats" }

func (this *cmdFormats) Description() string { return "Print a list of supported formats" }

func (this *cmdFormats) Usage() string {
	fs, _ := this.newFlagSet()
	return getFlagsUsage(fs)
}

func (this *cmdFormats) Run() (err error) {
	fmt.Printf("Supported formats:\n")
	for _, v := range getKnownDecoders() {
		fmt.Printf("    %v\n", v)
	}
	return
}

func (this *cmdFormats) Init(args []string) (err error) {
	return
}

func (this *cmdFormats) newFlagSet() (fs *flag.FlagSet, fields fieldValues) {
	fields = make(fieldValues)
	fs = flag.NewFlagSet("formats", flag.ContinueOnError)

	return
}

func init() {
	addCommand(new(cmdFormats))
}
