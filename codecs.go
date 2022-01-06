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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"

	"github.com/kstenerud/go-concise-encoding/ce"
	"github.com/kstenerud/go-concise-encoding/options"
)

type encoderConfig struct {
	indentSpaces int
}

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
	knownDecoders["json"] = decodeJSON
	knownEncoders["json"] = encodeJSON
	knownDecoders["xml"] = decodeXML
	knownEncoders["xml"] = encodeXML
	addCommand(new(cmdConvert))
}

type decoder func(io.Reader) (interface{}, error)
type encoder func(interface{}, io.Writer, *encoderConfig) error

func decodeCBE(reader io.Reader) (result interface{}, err error) {
	result, err = ce.UnmarshalCBE(reader, result, nil)
	return
}

func encodeCBE(value interface{}, writer io.Writer, config *encoderConfig) (err error) {
	err = ce.MarshalCBE(value, writer, nil)
	return
}

func decodeCTE(reader io.Reader) (result interface{}, err error) {
	result, err = ce.UnmarshalCTE(reader, result, nil)
	return
}

func encodeCTE(value interface{}, writer io.Writer, config *encoderConfig) (err error) {
	opts := options.DefaultCTEMarshalerOptions()
	opts.Encoder.Indent = strings.Repeat(" ", config.indentSpaces)
	err = ce.MarshalCTE(value, writer, opts)
	return
}

func decodeJSON(reader io.Reader) (result interface{}, err error) {
	document, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	v := make(map[string]interface{})

	err = json.Unmarshal(document, &v)
	result = v
	return
}

func encodeJSON(value interface{}, writer io.Writer, config *encoderConfig) (err error) {
	value = coerceToJSONable(value)
	var document []byte

	if config.indentSpaces == 0 {
		document, err = json.Marshal(value)
		if err != nil {
			return
		}
	} else {
		indent := strings.Repeat(" ", config.indentSpaces)
		document, err = json.MarshalIndent(value, "", indent)
		if err != nil {
			return
		}
	}

	_, err = writer.Write(document)
	return
}

// TODO: This needs tests
func coerceToJSONable(value interface{}) interface{} {
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() != reflect.String {
		newMap := make(map[string]interface{})
		iter := rv.MapRange()
		for iter.Next() {
			k := fmt.Sprintf("%v", iter.Key())
			v := coerceToJSONable(iter.Value().Interface())
			newMap[k] = v
		}
		value = newMap
	}
	return value
}

func decodeXML(reader io.Reader) (result interface{}, err error) {
	document, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	v := make(map[string]interface{})

	err = xml.Unmarshal(document, &v)
	result = v
	return
}

func encodeXML(value interface{}, writer io.Writer, config *encoderConfig) (err error) {
	value = coerceToXMLable(value)
	var document []byte

	if config.indentSpaces == 0 {
		document, err = xml.Marshal(value)
		if err != nil {
			return
		}
	} else {
		indent := strings.Repeat(" ", config.indentSpaces)
		document, err = xml.MarshalIndent(value, "", indent)
		if err != nil {
			return
		}
	}
	_, err = writer.Write(document)
	return
}

// TODO: This needs tests
func coerceToXMLable(value interface{}) interface{} {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return value
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.String {
			return value
		}
		newSlice := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			if elem.Kind() == reflect.String {
				newSlice = append(newSlice, elem.String())
			} else {
				newSlice = append(newSlice, fmt.Sprintf("%v", elem.Interface()))
			}
		}
		return newSlice
	case reflect.Map:
		if rv.Type().Key().Kind() == reflect.String && rv.Type().Elem().Kind() == reflect.String {
			return value
		}
		newMap := make(XMLStringMap)
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			if k.Kind() != reflect.String {
				k = reflect.ValueOf(fmt.Sprintf("%v", k.Interface()))
			}
			if v.Kind() != reflect.String {
				v = reflect.ValueOf(coerceToXMLable(v.Interface()))
			}
			newMap[k.String()] = v.String()
		}
		return newMap
	default:
		return fmt.Sprintf("%v", rv)
	}
}

type XMLStringMap map[string]string

func (_this XMLStringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	tokens := []xml.Token{start}

	for k, v := range _this {
		t := xml.StartElement{
			Name: xml.Name{
				Space: "",
				Local: k,
			},
		}
		tokens = append(tokens, t, xml.CharData(v), xml.EndElement{Name: t.Name})
	}

	tokens = append(tokens, xml.EndElement{Name: start.Name})

	for _, t := range tokens {
		if err = e.EncodeToken(t); err != nil {
			return
		}
	}

	return e.Flush()
}
