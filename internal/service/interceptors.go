package service

import (
	_ "github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors/cors"
	_ "github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors/http"
	_ "github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors/recover"
	_ "github.com/pjoc-team/pay-gateway/pkg/grpc/interceptors/tracing"
)
