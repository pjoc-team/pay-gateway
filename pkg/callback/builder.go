package callback

import (
	"bytes"
	"context"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"net/http"
)

func BuildChannelHttpRequest(ctx context.Context, r *http.Request) (
	request *pb.HTTPRequest, err error,
) {
	log := logger.ContextLog(ctx)

	var body []byte
	if body, err = GetBody(ctx, r); err != nil {
		return nil, err
	}
	request = &pb.HTTPRequest{}
	request.Body = body
	switch r.Method {
	case http.MethodGet:
		request.Method = pb.HTTPRequest_GET
	case http.MethodPost:
		request.Method = pb.HTTPRequest_POST
	default:
		log.Warnf("unknown http method: %v", r.Method)
		request.Method = pb.HTTPRequest_POST
	}
	request.Header = GetHeader(r)
	return
}

func GetHeader(r *http.Request) map[string]string {
	header := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}
	return header
}

func GetBody(ctx context.Context, r *http.Request) (data []byte, err error) {
	log := logger.ContextLog(ctx)

	body := r.Body
	defer func() {
		err2 := body.Close()
		if err2 != nil{
			log.Error(err2.Error())
			err = err2
		}
	}()
	buffer := bytes.Buffer{}
	if n, err := buffer.ReadFrom(body); err != nil {
		log.Errorf("Failed when read body! error: %v", err.Error())
		return nil, err
	} else {
		log.Debugf("Read byte size: %d", n)
		return buffer.Bytes(), nil
	}

}
