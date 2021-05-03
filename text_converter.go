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
)

type StringifyWriter struct {
	writer io.Writer
	buff   [2]byte
}

func newStringifyWriter(writer io.Writer) io.Writer {
	this := &StringifyWriter{
		writer: writer,
	}
	this.buff[0] = '\\'
	return this
}

func (this *StringifyWriter) Write(p []byte) (n int, err error) {
	var offset int
	for i := 0; i < len(p); i++ {
		this.buff[1] = p[i]
		switch this.buff[1] {
		case '"', '\\':
			offset = 0
		default:
			offset = 1
		}
		if _, err = this.writer.Write(this.buff[offset:]); err != nil {
			return
		}
		n++
	}
	return
}

type CWriter struct {
	writer      io.Writer
	isFirstbyte bool
	buff        [6]byte // , 0x00
}

func newCWriter(writer io.Writer) io.Writer {
	this := &CWriter{
		writer:      writer,
		isFirstbyte: true,
	}
	this.buff[0] = ','
	this.buff[1] = ' '
	this.buff[2] = '0'
	this.buff[3] = 'x'
	return this
}

func (this *CWriter) Write(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		b := p[i]
		this.buff[4] = hexChars[b>>4]
		this.buff[5] = hexChars[b&15]
		buff := this.buff[:]
		if this.isFirstbyte {
			buff = buff[2:]
			this.isFirstbyte = false
		}
		if _, err = this.writer.Write(buff); err != nil {
			return
		}
		n++
	}
	return
}

type HexWriter struct {
	writer      io.Writer
	isFirstbyte bool
	buff        [3]byte
}

var hexChars = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

func newHexWriter(writer io.Writer) io.Writer {
	this := &HexWriter{
		writer:      writer,
		isFirstbyte: true,
	}
	this.buff[0] = ' '
	return this
}

func (this *HexWriter) Write(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		b := p[i]
		this.buff[1] = hexChars[b>>4]
		this.buff[2] = hexChars[b&15]
		buff := this.buff[:]
		if this.isFirstbyte {
			buff = buff[1:]
			this.isFirstbyte = false
		}
		if _, err = this.writer.Write(buff); err != nil {
			return
		}
		n++
	}
	return
}

func newHexReader(reader io.Reader) io.Reader {
	return &HexReader{
		TokenReader{
			reader: reader,
		},
	}
}

type HexReader struct {
	TokenReader
}

func (this HexReader) convertOneByte() (b byte, err error) {
	var token []byte
	token, err = this.readToken()
	if err != nil {
		return
	}

	if len(token) != 2 {
		err = fmt.Errorf("Offset %v: Cannot convert %v to hex", this.offset, string(token))
		return
	}

	for i := 0; i < len(token); i++ {
		flags := charFlags[token[i]]
		if flags&charFlagHex == 0 {
			return 0, fmt.Errorf("Offset %v: Cannot convert %v to hex", this.offset, string(token))
		}
		b = (b << 4) | (flags & charValueMask)
	}
	return
}

func (this HexReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		var b byte
		b, err = this.convertOneByte()
		p[i] = b
		if err != nil {
			return
		}
		n++
	}
	return
}

// ============================================================================

func newTextByteReader(reader io.Reader) io.Reader {
	return &TextByteReader{
		TokenReader{
			reader: reader,
		},
	}
}

type TextByteReader struct {
	TokenReader
}

func (this TextByteReader) convertOneByte() (b byte, err error) {
	var token []byte
	token, err = this.readToken()
	if err != nil {
		return
	}

	if len(token) == 4 && token[0] == '0' && (token[1] == 'x' || token[1] == 'X') {
		for i := 2; i < len(token); i++ {
			flags := charFlags[token[i]]
			if flags&charFlagHex == 0 {
				return 0, fmt.Errorf("Offset %v: Cannot convert %v to hex", this.offset, string(token))
			}
			b = (b << 4) | (flags & charValueMask)
		}
		return
	}

	var v uint64
	for i := 0; i < len(token); i++ {
		flags := charFlags[token[i]]
		if flags&charFlagDecimal == 0 {
			return 0, fmt.Errorf("Offset %v: Cannot convert %v to decimal", this.offset, string(token))
		}
		v = v*10 + uint64(flags&charValueMask)
	}
	if v > 0xff {
		return 0, fmt.Errorf("Offset %v: Value %v is too big for a byte", this.offset, string(token))
	}

	b = byte(v)
	return
}

func (this TextByteReader) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		var b byte
		b, err = this.convertOneByte()
		p[i] = b
		if err != nil {
			return
		}
	}
	return
}

// ============================================================================

type TokenReader struct {
	reader io.Reader
	offset uint64
	buff   [1]byte
	token  []byte
}

func (this TokenReader) readByte() (b byte, err error) {
	_, err = this.reader.Read(this.buff[:])
	b = this.buff[0]
	this.offset++
	return
}

func (this TokenReader) readToken() (token []byte, err error) {
	this.token = this.token[:0]
	var b byte

	for {
		if b, err = this.readByte(); err != nil {
			return
		}
		if charFlags[b]&charFlagToken != 0 {
			break
		}
	}

	for {
		this.token = append(this.token, b)
		if b, err = this.readByte(); err != nil {
			break
		}
		if charFlags[b]&charFlagToken == 0 {
			break
		}
	}

	if err == io.EOF {
		err = nil
	}
	token = this.token
	return
}

// ============================================================================

var charFlags [256]byte

const (
	charFlagInvalid = 0x80
	charFlagHex     = 0x40
	charFlagDecimal = 0x20
	charFlagToken   = 0x10
	charValueMask   = 0x0f
)

func init() {
	for i := 0; i < len(charFlags); i++ {
		charFlags[i] = charFlagInvalid
	}
	for i := '0'; i <= '9'; i++ {
		charFlags[i] = byte(i-'0') | charFlagToken | charFlagDecimal | charFlagHex
	}
	for i := 'a'; i <= 'f'; i++ {
		charFlags[i] = byte(i-'a'+10) | charFlagToken | charFlagHex
	}
	for i := 'A'; i <= 'F'; i++ {
		charFlags[i] = byte(i-'A'+10) | charFlagToken | charFlagHex
	}
	charFlags['x'] |= charFlagToken
	charFlags['X'] |= charFlagToken
}
