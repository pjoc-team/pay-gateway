package sign

import (
	"bytes"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
)

// Type
type Type string

const (
	// TypeMd5 md5 sign
	TypeMd5 Type = "MD5"
	// TypeSha256WithRSA ras sign
	TypeSha256WithRSA Type = "RSA"
)

func init() {
	initEncodingMap()
}

var encodingMap = make(map[string]encoding.Encoding)

func initEncodingMap() {
	encodingMap["GBK"] = simplifiedchinese.GBK
	encodingMap["gbk"] = simplifiedchinese.GBK
}

func stringToBytes(str string, charset string) ([]byte, error) {
	bts := []byte(str)
	enc, exists := encodingMap[charset]
	if !exists {
		// Default utf-8
		return bts, nil
	}
	reader := transform.NewReader(bytes.NewReader(bts), enc.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return bts, nil
	}
	return d, nil
}

// GbkToUtf8 convert gbk to utf8
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// Utf8ToGbk convert utf8 to gbk
func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
