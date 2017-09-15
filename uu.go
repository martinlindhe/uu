package uu

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// ...
const (
	maxBytesPerLine = 45
	UUAlphabet      = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_"
)

var (
	encoding = base64.NewEncoding(UUAlphabet)
)

// Decoded holds result from uuencoded decode
type Decoded struct {
	Data     []byte
	Filename string
	Mode     string
}

// Decode decodes UUencoded text
func Decode(data []byte) (*Decoded, error) {
	rows := strings.Split(string(data), "\n")
	dec := &Decoded{}
	if string(strings.Split(rows[0], " ")[0]) != "begin" {
		return nil, fmt.Errorf("invalid format")
	}

	if string(strings.Split(rows[0], " ")[1]) == " " || string(strings.Split(rows[0], " ")[1]) == "" {
		return nil, fmt.Errorf("invalid file permissions")
	}
	dec.Mode = strings.Split(rows[0], " ")[1]

	if string(strings.Split(rows[0], " ")[2]) == " " || string(strings.Split(rows[0], " ")[2]) == "" {
		return nil, fmt.Errorf("invalid filename")
	}
	dec.Filename = strings.Split(rows[0], " ")[2]

	if string(rows[len(rows)-2]) != "end" {
		return nil, fmt.Errorf("invalid format: no 'end' marker found")
	}
	if string(rows[len(rows)-3]) != "`" {
		return nil, fmt.Errorf("invalid ending format")
	}

	rows = rows[1 : len(rows)-3]

	var err error
	dec.Data, err = DecodeBlock(rows)
	return dec, err
}

// DecodeBlock decodes a uuencoded text block
func DecodeBlock(rows []string) ([]byte, error) {
	data := []byte{}
	for i, row := range rows {
		res, err := DecodeLine(row)
		if err != nil {
			return data, fmt.Errorf("DecodeBlock at line %d: %s", i+1, err)
		}
		data = append(data, res...)
	}
	return data, nil
}

// DecodeLine decodes a single line of uuencoded text
func DecodeLine(s string) ([]byte, error) {
	// fix up non-standard padding `, to make golang's base64 not freak out
	s = strings.Replace(s, "`", " ", -1)

	data := []byte(s)
	l := data[0] - 32 // length
	res, err := encoding.DecodeString(s[1:])
	if err != nil {
		return res, err
	}
	return res[0:l], nil
}

// Encode encodes data into uuencoded format, with header and footer
func Encode(data []byte, filename, mode string) ([]byte, error) {
	out := []byte{}
	out = append(out, fmt.Sprintf("begin %s %s\n", mode, filename)...)

	enc, err := EncodeBlock(data)
	if err != nil {
		return nil, err
	}
	out = append(out, enc...)

	out = append(out, "`\nend\n"...)
	return out, nil
}

// EncodeBlock encodes data in raw uunecoded format
func EncodeBlock(data []byte) ([]byte, error) {
	out := []byte{}
	buf := bytes.NewBuffer(data)
	inputBlock := make([]byte, maxBytesPerLine)
	for {
		n, err := buf.Read(inputBlock)
		if n == 0 && err != nil {
			if err != io.EOF {
				return out, err
			}
			break
		}
		out = append(out, byte(n+32)) // length
		out = append(out, encoding.EncodeToString(inputBlock)...)
		out = append(out, '\n')
	}
	return out, nil
}
