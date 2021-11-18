package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/kstenerud/go-concise-encoding/cbe"
	"github.com/kstenerud/go-concise-encoding/cte"
	"github.com/kstenerud/go-concise-encoding/events"
	"github.com/kstenerud/go-concise-encoding/options"
	"github.com/kstenerud/go-concise-encoding/rules"
	"github.com/kstenerud/go-concise-encoding/version"

	"github.com/cockroachdb/apd/v2"
	compact_float "github.com/kstenerud/go-compact-float"
	compact_time "github.com/kstenerud/go-compact-time"
)

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

func CBEToXML(in io.Reader, out io.Writer, config *encoderConfig) error {
	eventReceiver := NewXMLEventReceiver(out, config.indentSpaces)
	decoder := cbe.NewDecoder(nil)
	return decoder.Decode(in, eventReceiver)
}

func CTEToXML(in io.Reader, out io.Writer, config *encoderConfig) error {
	eventReceiver := NewXMLEventReceiver(out, config.indentSpaces)
	decoder := cte.NewDecoder(nil)
	return decoder.Decode(in, eventReceiver)
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
			encoder.OnComment(true, elem)
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

type XMLEventReceiver struct {
	encoder             *xml.Encoder
	stack               []string
	attributes          []xml.Attr
	key                 xml.Name
	stage               MarkupStage
	stringBuffer        []byte
	arrayBytesRemaining uint64
	moreChunksComing    bool
}

func NewXMLEventReceiver(out io.Writer, indentSpaces int) *XMLEventReceiver {
	rcv := &XMLEventReceiver{
		encoder: xml.NewEncoder(out),
	}
	rcv.encoder.Indent("", generateIndentSpaces(indentSpaces))
	return rcv
}

func (_this *XMLEventReceiver) OnMarkup(name []byte) {
	_this.stackMarkup(string(name))
	_this.attributes = _this.attributes[:0]
	_this.stage = MarkupStageAttributeKey
}

func (_this *XMLEventReceiver) OnComment(isMultiline bool, contents []byte) {
	_this.encoder.EncodeToken(xml.Comment(string(contents)))
}

func (_this *XMLEventReceiver) OnStringlikeArray(arrayType events.ArrayType, str string) {
	if arrayType != events.ArrayTypeResourceID && arrayType != events.ArrayTypeString {
		panic(fmt.Errorf("Cannot convert array type %v to XML", arrayType))
	}

	switch _this.stage {
	case MarkupStageAttributeKey:
		_this.key = toXMLName(str)
		_this.stage = MarkupStageAttributeValue
	case MarkupStageAttributeValue:
		_this.attributes = append(_this.attributes, xml.Attr{
			Name:  _this.key,
			Value: str,
		})
		_this.stage = MarkupStageAttributeKey
	case MarkupStageContents:
		_this.encoder.EncodeToken(xml.CharData(str))
		_this.clearStringBuffer()
	default:
		panic(fmt.Errorf("Non-markup content detected"))
	}
}

func (_this *XMLEventReceiver) OnEnd() {
	switch _this.stage {
	case MarkupStageAttributeKey:
		name := _this.getCurrentMarkupName()
		_this.encoder.EncodeToken(xml.StartElement{
			Name: toXMLName(name),
			Attr: _this.attributes,
		})
		_this.stage = MarkupStageContents
	case MarkupStageContents:
		name := _this.getCurrentMarkupName()
		_this.encoder.EncodeToken(xml.EndElement{
			Name: toXMLName(name),
		})
		_this.unstackMarkup()
	default:
		panic(fmt.Errorf("BUG: Unhandled stage: %v", _this.stage))
	}
}

func (_this *XMLEventReceiver) stackMarkup(name string) {
	_this.stack = append(_this.stack, name)
}

func (_this *XMLEventReceiver) getCurrentMarkupName() string {
	return _this.stack[len(_this.stack)-1]
}

func (_this *XMLEventReceiver) unstackMarkup() {
	_this.stack = _this.stack[:len(_this.stack)-1]
}

func (_this *XMLEventReceiver) clearStringBuffer() {
	_this.stringBuffer = _this.stringBuffer[:0]
}

func (_this *XMLEventReceiver) appendStringBuffer(data []byte) {
	_this.stringBuffer = append(_this.stringBuffer, data...)
}

func (_this *XMLEventReceiver) OnBeginDocument() {}

func (_this *XMLEventReceiver) OnVersion(uint64) {
	_this.stage = MarkupStageNonMarkup
}

func (_this *XMLEventReceiver) OnPadding(int) {}

func (_this *XMLEventReceiver) OnNA() {
	panic("Cannot convert NA type to xml")
}

func (_this *XMLEventReceiver) OnNull() {
	_this.OnStringlikeArray(events.ArrayTypeString, "null")
}

func (_this *XMLEventReceiver) OnBool(v bool) {
	if v {
		_this.OnTrue()
	} else {
		_this.OnFalse()
	}
}

func (_this *XMLEventReceiver) OnTrue() {
	_this.OnStringlikeArray(events.ArrayTypeString, "true")
}

func (_this *XMLEventReceiver) OnFalse() {
	_this.OnStringlikeArray(events.ArrayTypeString, "false")
}

func (_this *XMLEventReceiver) OnPositiveInt(v uint64) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnNegativeInt(v uint64) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("-%v", v))
}

func (_this *XMLEventReceiver) OnInt(v int64) {
	if v < 0 {
		_this.OnNegativeInt(uint64(-v))
	} else {
		_this.OnPositiveInt(uint64(v))
	}
}

func (_this *XMLEventReceiver) OnBigInt(v *big.Int) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnFloat(v float64) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnBigFloat(v *big.Float) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnDecimalFloat(v compact_float.DFloat) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnBigDecimalFloat(v *apd.Decimal) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnNan(isSignaling bool) {
	if isSignaling {
		_this.OnStringlikeArray(events.ArrayTypeString, "snan")
	} else {
		_this.OnStringlikeArray(events.ArrayTypeString, "nan")
	}
}

func (_this *XMLEventReceiver) OnUID(v []byte) {
	_this.OnStringlikeArray(events.ArrayTypeString,
		fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x",
			v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], v[11], v[12], v[13], v[14], v[15]))

}

func (_this *XMLEventReceiver) OnTime(v time.Time) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnCompactTime(v compact_time.Time) {
	_this.OnStringlikeArray(events.ArrayTypeString, fmt.Sprintf("%v", v))
}

func (_this *XMLEventReceiver) OnArray(arrayType events.ArrayType, elemCount uint64, elems []byte) {
	switch arrayType {
	case events.ArrayTypeResourceID, events.ArrayTypeString:
		_this.OnStringlikeArray(arrayType, string(elems))
	default:
		panic(fmt.Errorf("Cannot convert array type %v to XML", arrayType))
	}
}

func (_this *XMLEventReceiver) OnArrayBegin(arrayType events.ArrayType) {
	if arrayType != events.ArrayTypeResourceID && arrayType != events.ArrayTypeString {
		panic(fmt.Errorf("Cannot convert array type %v to XML", arrayType))
	}
	_this.clearStringBuffer()
}

func (_this *XMLEventReceiver) OnArrayChunk(chunkSize uint64, moreChunksComing bool) {
	_this.arrayBytesRemaining = chunkSize
	_this.moreChunksComing = moreChunksComing
}

func (_this *XMLEventReceiver) OnArrayData(data []byte) {
	_this.appendStringBuffer(data)
	_this.arrayBytesRemaining -= uint64(len(data))
	if _this.arrayBytesRemaining == 0 && !_this.moreChunksComing {
		_this.OnStringlikeArray(events.ArrayTypeString, string(_this.stringBuffer))
	}
}

func (_this *XMLEventReceiver) OnList() {
	panic(fmt.Errorf("Cannot convert list to XML"))
}

func (_this *XMLEventReceiver) OnMap() {
	panic(fmt.Errorf("Cannot convert map to XML"))
}

func (_this *XMLEventReceiver) OnNode() {
	panic("Cannot convert node to xml")
}

func (_this *XMLEventReceiver) OnEdge() {
	panic(fmt.Errorf("Cannot convert edge to XML"))
}

func (_this *XMLEventReceiver) OnMarker([]byte) {
	panic(fmt.Errorf("Cannot convert marker to XML"))
}

func (_this *XMLEventReceiver) OnReference([]byte) {
	panic(fmt.Errorf("Cannot convert reference to XML"))
}

func (_this *XMLEventReceiver) OnRIDReference() {
	panic(fmt.Errorf("Cannot convert resource ID ref to XML"))
}

func (_this *XMLEventReceiver) OnConstant(_ []byte) {
	panic(fmt.Errorf("Cannot convert constant to XML"))
}

func (_this *XMLEventReceiver) OnEndDocument() {}

func generateIndentSpaces(count int) string {
	var buff []byte
	for i := 0; i < count; i++ {
		buff = append(buff, ' ')
	}
	return string(buff)
}

func toXMLName(name string) xml.Name {
	split := strings.SplitN(name, ":", 2)
	if len(split) == 2 {
		return xml.Name{
			Space: split[0],
			Local: split[1],
		}
	}
	return xml.Name{
		Local: name,
	}
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

type MarkupStage int

const (
	MarkupStageNonMarkup MarkupStage = iota
	MarkupStageAttributeKey
	MarkupStageAttributeValue
	MarkupStageContents
)
