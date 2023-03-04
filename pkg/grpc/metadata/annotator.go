package metadata

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

var (
	keysFromQuery = []string{"guid"}
)

// ParseHeaderAndQueryToMD parse headers and some query args to meta data.
func ParseHeaderAndQueryToMD(_ context.Context, req *http.Request) metadata.MD {
	m := make(metadata.MD)

	// Header
	for k, v := range req.Header {
		if len(v) > 0 {
			m.Set(k, v[0])
		}
	}

	// Query
	query := req.URL.Query()
	for _, key := range keysFromQuery {
		if v := query.Get(key); v != "" {
			m.Set(key, v)
		}
	}

	// Method, Path
	m.Set(httpMethod, req.Method)
	m.Set(httpPath, req.URL.Path)
	return m
}
