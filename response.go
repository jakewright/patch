package patch

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// Response represents the response from a request
type Response struct {
	*http.Response
}

// BodyBytes returns the body as a byte slice
func (r *Response) BodyBytes() ([]byte, error) {
	switch rc := r.Body.(type) {
	case *bufCloser:
		return rc.Bytes(), nil

	default:
		defer func() { _ = rc.Close() }()

		// Replace the response body with a bufCloser
		buf := &bufCloser{}
		r.Body = buf

		// Use a TeeReader to read the body while
		// simultaneously piping it into the buffer
		tr := io.TeeReader(rc, buf)
		return ioutil.ReadAll(tr)
	}
}

// BodyString returns the body as a string
func (r *Response) BodyString() (string, error) {
	b, err := r.BodyBytes()
	return string(b), err
}

type DecodeHook func(status int) interface{}

func On2xx(v interface{}) DecodeHook {
	return func(status int) interface{} {
		if status >= 200 && status < 300 {
			return v
		}

		return nil
	}
}

func On4xx(v interface{}) DecodeHook {
	return func(status int) interface{} {
		if status >= 400 && status < 500 {
			return v
		}

		return nil
	}
}

func On5xx(v interface{}) DecodeHook {
	return func(status int) interface{} {
		if status >= 500 && status < 600 {
			return v
		}

		return nil
	}
}

func OnNon2xx(v interface{}) DecodeHook {
	return func(status int) interface{} {
		if status >= 200 && status < 300 {
			return nil
		}

		return v
	}
}

func OnStatus(status int, v interface{}) DecodeHook {
	return func(s int) interface{} {
		if status == s {
			return v
		}

		return nil
	}
}

func (r *Response) Decode(targets ...interface{}) error {
	dec, err := inferDecoder(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	return r.DecodeUsing(dec, targets...)
}

func (r *Response) DecodeJSON(targets ...interface{}) error {
	return r.DecodeUsing(jsonDecoder, targets...)
}

// DecodeUsing decodes the response into the receivers using the given Decoder
func (r *Response) DecodeUsing(dec Decoder, targets ...interface{}) error {
	body, err := r.BodyBytes()
	if err != nil {
		return err
	}

	for _, receiver := range targets {
		switch v := receiver.(type) {
		case DecodeHook:
			receiver = v(r.StatusCode)
		}

		if receiver == nil {
			continue
		}

		if err := dec.Decode(body, receiver); err != nil {
			return err
		}
	}

	return nil
}

type bufCloser struct {
	bytes.Buffer
}

// Close is a no-op
func (b *bufCloser) Close() error {
	return nil
}
