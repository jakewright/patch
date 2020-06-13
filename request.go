package patch

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Request holds the information needed to make an HTTP request
type Request struct {
	Ctx     context.Context
	Method  string
	URL     string
	Headers http.Header
	Body    interface{}
	Encoder Encoder
}

func (r *Request) validate() error {
	switch {
	case !validMethod(r.Method):
		return InvalidMethodError(r.Method)
	}
	return nil
}

func (r *Request) prepareBody(defaultEncoder Encoder) (io.Reader, string, error) {
	if r.Body == nil {
		return nil, "", nil
	}

	enc := r.Encoder
	if enc == nil {
		enc = defaultEncoder
	}

	if enc == nil {
		return nil, "", fmt.Errorf("request has body but no encoder set on client or request")
	}

	reader, err := enc.Encode(r.Body)
	if err != nil {
		return nil, "", err
	}

	return reader, enc.ContentType(), nil
}

func validMethod(method string) bool {
	switch method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodPatch:
	case http.MethodDelete:
	case http.MethodConnect:
	case http.MethodOptions:
	case http.MethodTrace:
	default:
		return false
	}

	return true
}
