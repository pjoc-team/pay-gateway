package stream

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pjoc-team/pay-gateway/demo/gateway/proto"
	"google.golang.org/genproto/googleapis/api/httpbody"
)

type server struct {
}

func (s server) Download(ctx context.Context, e *empty.Empty) (*httpbody.HttpBody, error) {
	bd := &httpbody.HttpBody{
		ContentType: "text/html",
		Data:        []byte("Hello 2"),
	}
	return bd, nil
}

// NewStreamServer create server
func NewStreamServer() proto.StreamServiceServer {
	return &server{}
}

// func (s *server) Download(
// 	empty *empty.Empty, stream proto.StreamService_DownloadServer,
// ) error {
// 	msgs := []*httpbody.HttpBody{
// 		{
// 			ContentType: "text/html",
// 			Data:        []byte("Hello 1"),
// 		}, {
// 			ContentType: "text/html",
// 			Data:        []byte("Hello 2"),
// 		},
// 	}
//
// 	for _, msg := range msgs {
// 		if err := stream.Send(msg); err != nil {
// 			return err
// 		}
//
// 		time.Sleep(5 * time.Millisecond)
// 	}
//
// 	return nil
// }
