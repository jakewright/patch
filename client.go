package patch

import (
	"context"
	"net/http"
	"net/url"
)

// Client is an HTTP client that uses the BaseClient to send requests
type Client struct {
	BaseURL         string
	DefaultEncoder  Encoder
	StatusValidator func(int) bool
	BaseClient      Doer
}

// Doer executes HTTP requests. It is implemented by http.Client{}.
// You can wrap an http.Client{} in a custom Doer implementation
// to create middleware.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// New returns a new Client with sensible defaults.
// The defaults can be overridden by supplying Options.
func New(opts ...Option) *Client {
	return NewFromBaseClient(&http.Client{
		Timeout: DefaultTimeout,
	}, opts...)
}

// NewFromBaseClient returns a new Client that wraps BaseClient
func NewFromBaseClient(baseClient Doer, opts ...Option) *Client {
	c := &Client{
		DefaultEncoder:  &EncoderJSON{},
		StatusValidator: DefaultStatusValidator,
		BaseClient:      baseClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string, v interface{}) (*Response, error) {
	r := &Request{Ctx: ctx, Method: http.MethodGet, URL: url}
	rsp, err := c.Send(r).Response()
	if err != nil {
		return rsp, err
	}
	if v != nil {
		return rsp, rsp.Decode(v)
	}
	return rsp, nil
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, body interface{}, v interface{}) (*Response, error) {
	r := &Request{Ctx: ctx, Method: http.MethodPost, URL: url, Body: body}
	rsp, err := c.Send(r).Response()
	if err != nil {
		return rsp, err
	}
	if v != nil {
		return rsp, rsp.Decode(v)
	}
	return rsp, nil
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, url string, body interface{}, v interface{}) (*Response, error) {
	r := &Request{Ctx: ctx, Method: http.MethodPut, URL: url, Body: body}
	rsp, err := c.Send(r).Response()
	if err != nil {
		return rsp, err
	}
	if v != nil {
		return rsp, rsp.Decode(v)
	}
	return rsp, nil
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, url string, body interface{}, v interface{}) (*Response, error) {
	r := &Request{Ctx: ctx, Method: http.MethodPatch, URL: url, Body: body}
	rsp, err := c.Send(r).Response()
	if err != nil {
		return rsp, err
	}
	if v != nil {
		return rsp, rsp.Decode(v)
	}
	return rsp, nil
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string, body interface{}, v interface{}) (*Response, error) {
	r := &Request{Ctx: ctx, Method: http.MethodDelete, URL: url, Body: body}
	rsp, err := c.Send(r).Response()
	if err != nil {
		return rsp, err
	}
	if v != nil {
		return rsp, rsp.Decode(v)
	}
	return rsp, nil
}

// Send performs the HTTP request and returns a Future
func (c *Client) Send(request *Request) *Future {
	done := make(chan struct{})
	ftr := &Future{done: done}

	go func() {
		defer close(done)
		ftr.response, ftr.err = c.send(request)
	}()

	return ftr
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	rsp, err := c.BaseClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Execute the status validator if set
	if c.StatusValidator != nil && !c.StatusValidator(rsp.StatusCode) {
		return rsp, BadStatusError(rsp.StatusCode)
	}

	return rsp, nil
}

func (c *Client) send(request *Request) (*Response, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	/* Build the HTTP request */

	path := request.URL

	if c.BaseURL != "" {
		base, err := url.Parse(c.BaseURL)
		if err != nil {
			return nil, err
		}

		ref, err := url.Parse(request.URL)
		if err != nil {
			return nil, err
		}

		path = base.ResolveReference(ref).String()
	}

	body, contentType, err := request.prepareBody(c.DefaultEncoder)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(request.Method, path, body)
	if err != nil {
		return nil, err
	}

	if request.Ctx != nil {
		req = req.WithContext(request.Ctx)
	}

	if request.Headers != nil {
		req.Header = request.Headers
	}

	// Set the Content-Type header (unless an override was provided in request)
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	/* Make the HTTP request */

	rsp, err := c.Do(req)
	return &Response{Response: rsp}, err
}
