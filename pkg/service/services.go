package service

import (
	pb "github.com/pjoc-team/pay-proto/go"
)

// Discovery discovery server
type Discovery struct {
}

// GetChannelClient get channel client of id
func (d *Discovery) GetChannelClient(id string) (pb.PayChannelClient, error) {
	return nil, nil
}
