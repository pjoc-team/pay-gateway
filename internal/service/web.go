package service

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

// originContentType context key to mark origin ContentType header
type originContentType string

const (
	originContentTypeContextKey originContentType = "originContentType"

	rawWebContentType = "application/raw-web"
)


var (
	typeOfBytes    = reflect.TypeOf((*[]byte)(nil))
	typeOfHttpBody = reflect.TypeOf((*httpbody.HttpBody)(nil))
)

func rawWebOption(jsonPb *runtime.JSONPb) runtime.ServeMuxOption {
	return runtime.WithMarshalerOption(rawWebContentType, &rawJSONPb{JSONPb: jsonPb})
}

// customMimeWrapper process custom http path
// https://github.com/grpc-ecosystem/grpc-gateway/issues/652#issuecomment-428059210
func customMimeWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/web/") {
				contentType := r.Header.Get("Content-Type")
				nr := r.WithContext(
					context.WithValue(
						r.Context(), originContentTypeContextKey,
						contentType,
					),
				)
				r.Header.Set("Content-Type", rawWebContentType)
				h.ServeHTTP(w, nr)
				return
			}
			h.ServeHTTP(w, r)
		},
	)
}

type rawJSONPb struct {
	*runtime.JSONPb
}

func (*rawJSONPb) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			rawData, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			err = decode(rawData, v)
			return err
		},
	)
}

func decode(rawData []byte, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", v)
	}
	t := rv.Type()

	// rv = rv.Elem()
	// tp := t
	switch t {
	case typeOfHttpBody:
		// hb := v.(*httpbody.HttpBody)
		hb := &httpbody.HttpBody{}
		hb.Data = rawData
		rv.Elem().Set(reflect.ValueOf(hb).Elem())
	case typeOfBytes:
		// bts := v.([]byte)
		// for _, datum := range rawData {
		// 	bts = append(bts, datum)
		// }
		// copy(bts[:], rawData[:])
		// rv.Set(reflect.ValueOf(rawData))
		rv.Elem().Set(reflect.ValueOf(rawData))
	default:
		return fmt.Errorf("type must be []byte or *httpbody.HttpBody but got %T", v)
	}

	return nil
}
