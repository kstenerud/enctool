package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/kstenerud/go-concise-encoding/cbe"
	"github.com/kstenerud/go-concise-encoding/configuration"
	"github.com/kstenerud/go-concise-encoding/cte"
	"github.com/kstenerud/go-concise-encoding/rules"
	qrcode "github.com/kstenerud/go-qrcode"
	"github.com/liyue201/goqr"
)

type converter func(io.Reader, io.Writer, *encoderConfig) error

var knownConverters = map[string]converter{
	"cbe-cbe":   CBEToCBE,
	"cbe-cte":   CBEToCTE,
	"cbe-json":  CBEToJSON,
	"cte-cte":   CTEToCTE,
	"cte-cbe":   CTEToCBE,
	"cbe-qr":    CBEToQR,
	"cte-qr":    CTEToQR,
	"cte-qrt":   CTEToQRT,
	"qr-cte":    QRToCTE,
	"qr-cbe":    QRToCBE,
	"cte-json":  CTEToJSON,
	"json-json": JSONToJSON,
	"json-cbe":  JSONToCBE,
	"json-cte":  JSONToCTE,
	"xml-cbe":   XMLToCBE,
	"xml-cte":   XMLToCTE,
	"cbe-xml":   CBEToXML,
	"cte-xml":   CTEToXML,
}

func getConverter(id string) (converter, error) {
	converter := knownConverters[id]
	if converter == nil {
		return nil, fmt.Errorf("%v: Unknown converter", id)
	}
	return converter, nil
}

func CBEToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	encoder := cbe.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cbe.NewDecoder(opts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CBEToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	opts.Encoder.CTE.Indent = generateSpaces(config.indentSpaces)
	encoder := cte.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cbe.NewDecoder(opts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CTEToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	encoder := cte.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cte.NewDecoder(opts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CTEToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	encoder := cbe.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cte.NewDecoder(opts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(in, rules)
}

func CBEToQR(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	encoder := cbe.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cbe.NewDecoder(opts)

	buff := &bytes.Buffer{}
	encoder.PrepareToEncode(buff)
	err := decoder.Decode(in, rules)
	if err != nil {
		return err
	}

	q, err := qrcode.New(buff.Bytes(), qrcode.RecoveryLevel(config.errorCorrection))
	if err != nil {
		return err
	}
	q.BorderSize = int(config.borderSize)
	png, err := q.PNG(int(config.imageSize))
	if err != nil {
		return err
	}

	_, err = out.Write(png)
	return err
}

func CTEToQR(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	encoder := cbe.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cte.NewDecoder(opts)

	buff := &bytes.Buffer{}
	encoder.PrepareToEncode(buff)
	err := decoder.Decode(in, rules)
	if err != nil {
		return err
	}

	q, err := qrcode.New(buff.Bytes(), qrcode.RecoveryLevel(config.errorCorrection))
	if err != nil {
		return err
	}
	q.BorderSize = int(config.borderSize)
	png, err := q.PNG(int(config.imageSize))
	if err != nil {
		return err
	}

	_, err = out.Write(png)
	return err
}

func CTEToQRT(in io.Reader, out io.Writer, config *encoderConfig) error {
	opts := configuration.New()
	encoder := cbe.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cte.NewDecoder(opts)

	buff := &bytes.Buffer{}
	encoder.PrepareToEncode(buff)
	err := decoder.Decode(in, rules)
	if err != nil {
		return err
	}

	q, err := qrcode.New(buff.Bytes(), qrcode.RecoveryLevel(config.errorCorrection))
	if err != nil {
		return err
	}
	if config.invertText {
		q.BorderSize = 0
	}
	art := q.ToString(config.invertText)
	_, err = out.Write([]byte(art))
	return err
}

func QRToCTE(in io.Reader, out io.Writer, config *encoderConfig) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return err
	}
	buff := &bytes.Buffer{}
	for _, qrCode := range qrCodes {
		buff.Write(qrCode.Payload)
	}

	opts := configuration.New()
	encoder := cte.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cbe.NewDecoder(opts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(buff, rules)
}

func QRToCBE(in io.Reader, out io.Writer, config *encoderConfig) error {
	img, _, err := image.Decode(in)
	if err != nil {
		return err
	}
	qrCodes, err := goqr.Recognize(img)
	if err != nil {
		return err
	}
	buff := &bytes.Buffer{}
	for _, qrCode := range qrCodes {
		buff.Write(qrCode.Payload)
	}

	opts := configuration.New()
	encoder := cbe.NewEncoder(opts)
	rules := rules.NewRules(encoder, opts)
	decoder := cbe.NewDecoder(opts)
	encoder.PrepareToEncode(out)
	return decoder.Decode(buff, rules)
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
