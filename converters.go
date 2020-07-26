package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/kstenerud/go-concise-encoding/options"

	"github.com/kstenerud/go-concise-encoding/cbe"
	"github.com/kstenerud/go-concise-encoding/cte"
	"github.com/kstenerud/go-concise-encoding/rules"
)

type converter func(io.Reader, io.Writer, *encoderConfig) error

var knownConverters = map[string]converter{
	"cbe-cte":  CBEToCTE,
	"cbe-json": CBEToJSON,
	"cte-cbe":  CTEToCBE,
	"cte-json": CTEToJSON,
	"json-cbe": JSONToCBE,
	"json-cte": JSONToCTE,
}

func getConverter(id string) (converter, error) {
	converter := knownConverters[id]
	if converter == nil {
		return nil, fmt.Errorf("%v: Unknown converter", id)
	}
	return converter, nil
}

func CBEToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	decoderOpts := options.DefaultCBEDecoderOptions()
	encoderOpts := options.DefaultCTEEncoderOptions()
	encoderOpts.Indent = generateSpaces(config.indentSpaces)
	rulesOpts := options.DefaultRuleOptions()
	encoder := cte.NewEncoder(out, encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	decoder := cbe.NewDecoder(in, rules, decoderOpts)
	return decoder.Decode()
}

func CTEToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	decoderOpts := options.DefaultCTEDecoderOptions()
	encoderOpts := options.DefaultCBEEncoderOptions()
	rulesOpts := options.DefaultRuleOptions()
	encoder := cbe.NewEncoder(out, encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	decoder := cte.NewDecoder(in, rules, decoderOpts)
	return decoder.Decode()
}

func JSONToCBE(in io.Reader, out io.Writer, config *encoderConfig) (err error) {
	object, err := decodeJSON(in)
	if err != nil {
		return
	}
	err = encodeCBE(object, out, config)
	return
}

func JSONToCTE(in io.Reader, out io.Writer, config *encoderConfig) (err error) {
	object, err := decodeJSON(in)
	if err != nil {
		return
	}
	err = encodeCTE(object, out, config)
	return
}

func CBEToJSON(in io.Reader, out io.Writer, config *encoderConfig) (err error) {
	object, err := decodeCBE(in)
	if err != nil {
		return
	}
	err = encodeJSON(object, out, config)
	return
}

func CTEToJSON(in io.Reader, out io.Writer, config *encoderConfig) (err error) {
	object, err := decodeCTE(in)
	if err != nil {
		return
	}
	err = encodeJSON(object, out, config)
	return
}

func generateSpaces(count int) string {
	var b strings.Builder
	for i := 0; i < count; i++ {
		b.WriteByte(' ')
	}
	return b.String()
}
