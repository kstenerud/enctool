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
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/kstenerud/go-concise-encoding"
)

func getKnownEncoders() []string {
	keys := make([]string, 0, len(knownEncoders))
	for k := range knownEncoders {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getKnownDecoders() []string {
	keys := make([]string, 0, len(knownDecoders))
	for k := range knownDecoders {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getDecoder(id string) (decoder, error) {
	decoder := knownDecoders[id]
	if decoder == nil {
		return nil, fmt.Errorf("%v: Unknown decoder", id)
	}
	return decoder, nil
}

func getEncoder(id string) (encoder, error) {
	encoder := knownEncoders[id]
	if encoder == nil {
		return nil, fmt.Errorf("%v: Unknown encoder", id)
	}
	return encoder, nil
}

var knownDecoders = make(map[string]decoder)
var knownEncoders = make(map[string]encoder)

func init() {
	knownDecoders["cbe"] = decodeCBE
	knownEncoders["cbe"] = encodeCBE
	knownDecoders["cte"] = decodeCTE
	knownEncoders["cte"] = encodeCTE
	addCommand(new(cmdConvert))
}

type decoder func(io.Reader) (interface{}, error)
type encoder func(interface{}, io.Writer) error

func decodeCBE(reader io.Reader) (result interface{}, err error) {
	document, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	result, err = concise_encoding.UnmarshalCBE(document, result, nil)
	return
}

func decodeCTE(reader io.Reader) (result interface{}, err error) {
	document, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	result, err = concise_encoding.UnmarshalCTE(document, result, nil)
	return
}

func encodeCBE(value interface{}, writer io.Writer) (err error) {
	options := &concise_encoding.CBEMarshalerOptions{
		Iterator: concise_encoding.IteratorOptions{
			UseReferences: true,
		},
	}
	document, err := concise_encoding.MarshalCBE(value, options)
	if err != nil {
		return
	}
	_, err = writer.Write(document)
	return
}

func encodeCTE(value interface{}, writer io.Writer) (err error) {
	options := &concise_encoding.CTEMarshalerOptions{
		Iterator: concise_encoding.IteratorOptions{
			UseReferences: true,
		},
	}
	document, err := concise_encoding.MarshalCTE(value, options)
	if err != nil {
		return
	}
	_, err = writer.Write(document)
	return
}
