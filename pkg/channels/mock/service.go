package mock

import (
	"context"
	"github.com/pjoc-team/pay-gateway/pkg/config"
	"github.com/pjoc-team/pay-gateway/pkg/generator"
	pb "github.com/pjoc-team/pay-proto/go"
	"github.com/pjoc-team/tracing/logger"
)

// server mock channel server
type server struct {
	cs config.Server
	g  *generator.Generator
}


// ChannelConfig channel config struct
type ChannelConfig struct {
	PublicKey  string `json:"public_key" yaml:"publicKey" validate:"required"`
	PrivateKey string `json:"private_key" yaml:"privateKey" validate:"required"`
}

// NewServer create mock channel server
func NewServer(cs config.Server) (pb.PayChannelServer, error) {
	g := generator.New("1", 1_000_000)

	ps := &server{
		cs: cs,
		g:  g,
	}
	return ps, nil
}

func (s *server) Pay(ctx context.Context, request *pb.ChannelPayRequest) (
	*pb.ChannelPayResponse, error,
) {
	log := logger.ContextLog(ctx)

	log.Warnf("receive mock pay request: %#v", request)

	resp := &pb.ChannelPayResponse{}
	resp.ChannelOrderId = s.g.GenerateID()
	return resp, nil
}

func (s *server) ChannelNotify(
	ctx context.Context, request *pb.ChannelNotifyRequest,
) (*pb.ChannelNotifyResponse, error) {
	log := logger.ContextLog(ctx)
	log.Warnf("receive mock pay request: %#v", request)

	resp := &pb.ChannelNotifyResponse{}
	resp.Status = pb.PayStatus_SUCCESS
	return resp, nil
}