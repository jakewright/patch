package patch

import (
	"fmt"
	"net/http"
	"time"
)

// DefaultTimeout is the default time limit for requests made by the client.
const DefaultTimeout = 30 * time.Second

var DefaultResponseInterceptor = StatusValidatorInterceptor(Accept2xx)

type Option func(c *Client)

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.BaseURL = url
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		switch bc := c.BaseClient.(type) {
		case *http.Client:
			bc.Timeout = d
			return
		}

		panic(fmt.Errorf("cannot set timeout on base client of type %T", c))
	}
}

func WithEncoder(enc Encoder) Option {
	return func(c *Client) {
		c.DefaultEncoder = enc
	}
}

func WithRequestInterceptor(f func(*http.Request) (*http.Request, error)) Option {
	return func(c *Client) {
		c.RequestInterceptor = f
	}
}

func WithResponseInterceptor(f func(*http.Response) (*http.Response, error)) Option {
	return func(c *Client) {
		c.ResponseInterceptor = f
	}
}

func StatusValidatorInterceptor(f func(status int) bool) func(*http.Response) (*http.Response, error) {
	return func(rsp *http.Response) (*http.Response, error) {
		if !f(rsp.StatusCode) {
			return rsp, BadStatusError(rsp.StatusCode)
		}
		return rsp, nil
	}
}

var Accept2xx = func(status int) bool {
	return status >= 200 && status < 300
}
