package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/kstenerud/go-concise-encoding/options"

	"github.com/kstenerud/go-concise-encoding/cbe"
	"github.com/kstenerud/go-concise-encoding/cte"
	"github.com/kstenerud/go-concise-encoding/events"
	"github.com/kstenerud/go-concise-encoding/rules"
	"github.com/kstenerud/go-concise-encoding/version"
)

type converter func(io.Reader, io.Writer, *encoderConfig) error

var knownConverters = map[string]converter{
	"cbe-cbe":   CBEToCBE,
	"cbe-cte":   CBEToCTE,
	"cbe-json":  CBEToJSON,
	"cte-cte":   CTEToCTE,
	"cte-cbe":   CTEToCBE,
	"cte-json":  CTEToJSON,
	"json-json": JSONToJSON,
	"json-cbe":  JSONToCBE,
	"json-cte":  JSONToCTE,
	"xml-cbe":   XMLToCBE,
	"xml-cte":   XMLToCTE,
}

func getConverter(id string) (converter, error) {
	converter := knownConverters[id]
	if converter == nil {
		return nil, fmt.Errorf("%v: Unknown converter", id)
	}
	return converter, nil
}

func CBEToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	decoderOpts := options.DefaultCBEDecoderOptions()
	encoderOpts := options.DefaultCBEEncoderOptions()
	rulesOpts := options.DefaultRuleOptions()
	encoder := cbe.NewEncoder(encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	decoder := cbe.NewDecoder(decoderOpts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CBEToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	decoderOpts := options.DefaultCBEDecoderOptions()
	encoderOpts := options.DefaultCTEEncoderOptions()
	encoderOpts.Indent = generateSpaces(config.indentSpaces)
	rulesOpts := options.DefaultRuleOptions()
	encoder := cte.NewEncoder(encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	decoder := cbe.NewDecoder(decoderOpts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CTEToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	decoderOpts := options.DefaultCTEDecoderOptions()
	encoderOpts := options.DefaultCTEEncoderOptions()
	rulesOpts := options.DefaultRuleOptions()
	encoder := cte.NewEncoder(encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	decoder := cte.NewDecoder(decoderOpts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CTEToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	decoderOpts := options.DefaultCTEDecoderOptions()
	encoderOpts := options.DefaultCBEEncoderOptions()
	rulesOpts := options.DefaultRuleOptions()
	encoder := cbe.NewEncoder(encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	decoder := cte.NewDecoder(decoderOpts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func JSONToCBE(in io.Reader, out io.Writer, config *encoderConfig) (err error) {
	object, err := decodeJSON(in)
	if err != nil {
		return
	}
	err = encodeCBE(object, out, config)
	return
}

func JSONToJSON(in io.Reader, out io.Writer, config *encoderConfig) (err error) {
	object, err := decodeJSON(in)
	if err != nil {
		return
	}
	err = encodeJSON(object, out, config)
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

func getMarkupNameBytes(name xml.Name) (nameBytes []byte) {
	if len(name.Space) > 0 {
		nameBytes = []byte(name.Space)
	}
	if len(name.Local) > 0 {
		nameBytes = append(nameBytes, []byte(name.Local)...)
	}
	return
}

func XMLToCE(in io.Reader, encoder events.DataEventReceiver) error {
	var token xml.Token
	var err error

	encoder.OnBeginDocument()
	encoder.OnVersion(version.ConciseEncodingVersion)

	decoder := xml.NewDecoder(in)
	for {
		token, err = decoder.Token()
		if err != nil {
			if err == io.EOF {
				encoder.OnEndDocument()
				return nil
			}
			return err
		}
		if _, ok := token.(xml.StartElement); ok {
			break
		}
	}

	for {
		switch elem := token.(type) {
		case xml.StartElement:
			encoder.OnMarkup(getMarkupNameBytes(elem.Name))
			for _, v := range elem.Attr {
				b := getMarkupNameBytes(v.Name)
				encoder.OnArray(events.ArrayTypeString, uint64(len(b)), b)
				encoder.OnStringlikeArray(events.ArrayTypeString, v.Value)
			}
			encoder.OnEnd()
		case xml.EndElement:
			encoder.OnEnd()
		case xml.CharData:
			str := strings.TrimSpace(string(elem))
			if len(str) > 0 {
				encoder.OnStringlikeArray(events.ArrayTypeString, str)
			}
		case xml.Comment:
			str := strings.TrimSpace(string(elem))
			if len(str) > 0 {
				encoder.OnComment()
				encoder.OnStringlikeArray(events.ArrayTypeString, str)
				encoder.OnEnd()
			}
		case xml.ProcInst:
			// TODO: Anything?
		case xml.Directive:
			// TODO: Anything?
		}

		token, err = decoder.Token()
		if err != nil {
			if err == io.EOF {
				encoder.OnEndDocument()
				return nil
			}
			return err
		}
	}

	return nil
}

func XMLToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	encoderOpts := options.DefaultCBEEncoderOptions()
	rulesOpts := options.DefaultRuleOptions()
	encoder := cbe.NewEncoder(encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	encoder.PrepareToEncode(out)

	return XMLToCE(in, rules)
}

func XMLToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	encoderOpts := options.DefaultCTEEncoderOptions()
	rulesOpts := options.DefaultRuleOptions()
	encoder := cte.NewEncoder(encoderOpts)
	rules := rules.NewRules(encoder, rulesOpts)
	encoder.PrepareToEncode(out)

	return XMLToCE(in, rules)
}
