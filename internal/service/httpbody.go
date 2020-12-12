package service

import (
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"io/ioutil"
)

type httpBodyUnmarshaler struct {
	runtime.Marshaler
}

func (h *httpBodyUnmarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			switch vt := v.(type) {
			case *httpbody.HttpBody:
				return h.decodeHttpBody(r, vt)
			case **httpbody.HttpBody:
				*vt = &httpbody.HttpBody{}
				return h.decodeHttpBody(r, *vt)
			default:
				return h.Marshaler.NewDecoder(r).Decode(v)
			}

		},
	)
}

func (h *httpBodyUnmarshaler) decodeHttpBody(r io.Reader, body *httpbody.HttpBody) error {
	rawData, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	body.Data = rawData
	return nil
}

func (h *httpBodyUnmarshaler) decode(rawData []byte, v interface{}) error {
	switch vt := v.(type) {
	case *httpbody.HttpBody:
		vt.Data = rawData
	case **httpbody.HttpBody:
		*vt = &httpbody.HttpBody{
			Data: rawData,
		}
	default:
		return fmt.Errorf("type must be []byte or *httpbody.HttpBody but got %T", v)
	}

	// rv := reflect.ValueOf(v)
	//
	// if rv.Kind() != reflect.Ptr {
	// 	return fmt.Errorf("%T is not a pointer", v)
	// }
	// t := rv.Type()
	//
	// // rv = rv.Elem()
	// // tp := t
	// switch t {
	// case typeOfHttpBody:
	// 	// hb := v.(*httpbody.HttpBody)
	// 	hb := &httpbody.HttpBody{}
	// 	hb.Data = rawData
	// 	rv.Elem().Set(reflect.ValueOf(hb).Elem())
	// case typeOfHttpBodyPtr:
	// 	hb := &httpbody.HttpBody{}
	// 	hb.Data = rawData
	// 	rv.Elem().Set(reflect.ValueOf(hb))
	// 	// rv.Elem().Elem().Set(reflect.ValueOf(hb).Elem())
	// case typeOfBytes:
	// 	// bts := v.([]byte)
	// 	// for _, datum := range rawData {
	// 	// 	bts = append(bts, datum)
	// 	// }
	// 	// copy(bts[:], rawData[:])
	// 	// rv.Set(reflect.ValueOf(rawData))
	// 	rv.Elem().Set(reflect.ValueOf(rawData))
	// default:
	// 	return fmt.Errorf("type must be []byte or *httpbody.HttpBody but got %T", v)
	// }

	return nil
}

func httpBodyOption() runtime.ServeMuxOption {
	jsonPb := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			Multiline:       false,
			Indent:          "",
			AllowPartial:    false,
			UseProtoNames:   true,
			UseEnumNumbers:  false,
			EmitUnpopulated: true,
			Resolver:        nil,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	internalMarshaler := &runtime.HTTPBodyMarshaler{
		// https://grpc-ecosystem.github.io/grpc-gateway/docs/development/v2-migration/
		Marshaler: jsonPb,
	}
	m := &httpBodyUnmarshaler{
		Marshaler: internalMarshaler,
	}

	marshalOpt := runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		m,
	)
	return marshalOpt
}
