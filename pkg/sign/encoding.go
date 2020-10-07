package sign

import (
	"bytes"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
)

const (
	SIGN_TYPE_MD5             = "MD5"
	SIGN_TYPE_SHA256_WITH_RSA = "RSA"
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
	if enc, exists := encodingMap[charset]; !exists {
		// Default utf-8
		return bts, nil
	} else {
		reader := transform.NewReader(bytes.NewReader(bts), enc.NewEncoder())
		d, e := ioutil.ReadAll(reader)
		if e != nil {
			return bts, nil
		}
		return d, nil
	}
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
