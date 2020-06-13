package patch

import (
	"fmt"
	"net/http"
	"time"
)

// DefaultTimeout is the default time limit for requests made by the client.
const DefaultTimeout = 30 * time.Second

// DefaultStatusValidator returns true for 2xx statuses, otherwise false.
var DefaultStatusValidator = func(status int) bool {
	return status >= 200 && status < 300
}

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

func WithStatusValidator(f func(int) bool) Option {
	return func(c *Client) {
		c.StatusValidator = f
	}
}

func WithEncoder(enc Encoder) Option {
	return func(c *Client) {
		c.DefaultEncoder = enc
	}
}
