package metadata

import (
	"context"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	md "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	remoteIPFromPeer = "remote-ip-from-peer"
	httpMethod       = "method"
	httpPath         = "path"
)

// MetaData simple meta data.
type MetaData map[string]string

// GrpcGatewayHeaders get headers from grpc gateway
func GrpcGatewayHeaders(ctx context.Context) (md.MD, bool) {
	rs := md.MD{}
	metadata, ok := md.FromIncomingContext(ctx)
	if !ok {
		return nil, ok
	}

	for k, v := range metadata {
		if strings.HasPrefix(k, runtime.MetadataPrefix) {
			rs[strings.TrimPrefix(k, runtime.MetadataPrefix)] = v
		}
	}
	return rs, ok
}

// FromIncomingContext parse meta data from ctx.
func FromIncomingContext(ctx context.Context) MetaData {
	m := make(MetaData)
	data, ok := md.FromIncomingContext(ctx)
	if !ok {
		return m
	}
	for k, v := range data {
		if len(v) == 0 {
			continue
		}
		k = strings.ReplaceAll(strings.ToLower(k), "_", "-")
		m[k] = v[0]
	}

	// get remote address from peer
	pr, ok := peer.FromContext(ctx)
	if ok && pr.Addr != nil {
		ss := strings.Split(pr.Addr.String(), ":")
		m[remoteIPFromPeer] = ss[0]
	}

	return m
}

// GetAuthorization get authorization.
func (m MetaData) GetAuthorization() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-authorization", "authorization"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetUserAgent get user-agent.
func (m MetaData) GetUserAgent() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-user-agent", "user-agent"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetReferer get referer.
func (m MetaData) GetReferer() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-referer", "referer"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetRealIP get real ip from x-real-ip or x-forwarded-for or remote ip from peer.
func (m MetaData) GetRealIP() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-x-real-ip", "x-real-ip"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	if ip := m["x-forwarded-for"]; ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return m.GetRemoteIP()
}

// GetRemoteIP get remote ip from x-forwarded-for or remote ip from peer.
func (m MetaData) GetRemoteIP() string {
	if len(m) == 0 {
		return ""
	}
	if ip := m["x-forwarded-for"]; ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return m[remoteIPFromPeer]
}

// GetDeviceID get device id from x-device-id. `device_id` passed by url query.
func (m MetaData) GetDeviceID() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-x-device-id", "x-device-id", "device-id"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetCaptchaToken get captcha token. `captcha_token` passed by url query.
func (m MetaData) GetCaptchaToken() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-x-captcha-token", "x-captcha-token", "captcha-token"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetProjectID get project id from x-project-id.
func (m MetaData) GetProjectID() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"grpcgateway-x-project-id", "x-project-id"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetHost get host from x-forwarded-host or host.
func (m MetaData) GetHost() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"x-forwarded-host", "host"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetGUID get guid from x-guid or guid.
func (m MetaData) GetGUID() string {
	if len(m) == 0 {
		return ""
	}
	for _, k := range []string{"x-guid", "guid"} {
		if v := m[k]; v != "" {
			return v
		}
	}
	return ""
}

// GetRequestID get request id from x-request-id.
func (m MetaData) GetRequestID() string {
	if len(m) == 0 {
		return ""
	}
	return m["x-request-id"]
}

// GetHTTPMethod get http method from http_method.
func (m MetaData) GetHTTPMethod() string {
	if len(m) == 0 {
		return ""
	}
	return m[httpMethod]
}

// GetHTTPPath get http path from http_path.
func (m MetaData) GetHTTPPath() string {
	if len(m) == 0 {
		return ""
	}
	return m[httpPath]
}
