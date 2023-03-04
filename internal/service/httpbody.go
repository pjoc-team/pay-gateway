package service

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/encoding/protojson"
)

type httpBodyUnmarshaler struct {
	runtime.Marshaler
}

func (h *httpBodyUnmarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			switch vt := v.(type) {
			case *httpbody.HttpBody:
				return h.decodeHTTPBody(r, vt)
			case **httpbody.HttpBody:
				*vt = &httpbody.HttpBody{}
				return h.decodeHTTPBody(r, *vt)
			default:
				return h.Marshaler.NewDecoder(r).Decode(v)
			}

		},
	)
}

func (h *httpBodyUnmarshaler) decodeHTTPBody(r io.Reader, body *httpbody.HttpBody) error {
	rawData, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	body.Data = rawData
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
