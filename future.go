package patch

// Future represents an in-flight request
type Future struct {
	done     <-chan struct{}
	response *Response
	err      error
}

// Response blocks until the response is available
func (f *Future) Response() (*Response, error) {
	<-f.done
	return f.response, f.err
}
