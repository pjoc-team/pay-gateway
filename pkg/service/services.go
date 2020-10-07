package service

import (
	pb "github.com/pjoc-team/pay-proto/go"
)

type Discovery struct {
}

func (d *Discovery) GetChannelClient(id string) (pb.PayChannelClient, error) {
	return nil, nil
}
