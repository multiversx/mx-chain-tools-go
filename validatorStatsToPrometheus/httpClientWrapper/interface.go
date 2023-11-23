package httpClientWrapper

import "context"

// HttpClient defines the behavior of http client able to make http requests
type HttpClient interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
	PostHTTP(ctx context.Context, endpoint string, data []byte) ([]byte, int, error)
	IsInterfaceNil() bool
}
