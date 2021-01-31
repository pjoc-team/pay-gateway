package callback

import (
	"bytes"
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/metadata"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
	"net/http"
)

// BuildChannelHTTPRequest build callback request
func BuildChannelHTTPRequest(ctx context.Context, r *http.Request) (
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

// GetHeader get http header
func GetHeader(r *http.Request) map[string]string {
	header := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			header[k] = v[0]
		}
	}
	return header
}

// GetBody get http body
func GetBody(ctx context.Context, r *http.Request) (data []byte, err error) {
	log := logger.ContextLog(ctx)

	body := r.Body
	defer func() {
		err2 := body.Close()
		if err2 != nil {
			log.Error(err2.Error())
			err = err2
		}
	}()
	buffer := bytes.Buffer{}
	n, err := buffer.ReadFrom(body)
	if err != nil {
		log.Errorf("Failed when read body! error: %v", err.Error())
		return nil, err
	}
	log.Debugf("Read byte size: %d", n)
	return buffer.Bytes(), nil

}

// BuildChannelRequest build channel notify request
func BuildChannelRequest(
	ctx context.Context, request *pb.HttpCallbackRequest, stream pb.ChannelCallback_CallbackByPostServer,
) (*pb.HTTPRequest, error) {
	log := logger.ContextLog(ctx)

	rs := &pb.HTTPRequest{
	}

	// method
	m, ok := pb.HTTPRequest_HttpMethod_value[request.HttpMethod]
	if ok {
		rs.Method = pb.HTTPRequest_HttpMethod(m)
	}

	// header
	headers, ok := metadata.GrpcGatewayHeaders(stream.Context())
	if ok {
		log.Debugf("headers: %v", headers)
		rs.Header = make(map[string]string)
		for k, v := range headers {
			rs.Header[k] = v[0]
		}
	}

	// body
	if request.Body != nil {
		rs.Body = request.Body.Data
	}

	md := metadata.FromIncomingContext(ctx)
	// rs.Url = md.GetHTTPPath() + "?" + md.GetHTTPRawQuery()
	rs.Url = md.GetHTTPURL()

	return rs, nil
}