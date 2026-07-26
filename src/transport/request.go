package transport

// Request represents an HTTP request to be sent.
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// NewRequest creates a new Request.
func NewRequest(method, path string) *Request {
	return &Request{
		Method:  method,
		Path:    path,
		Headers: make(map[string]string),
	}
}

// WithBody sets the request body.
func (r *Request) WithBody(body interface{}) *Request {
	r.Body = body
	return r
}

// WithHeader sets a request header.
func (r *Request) WithHeader(key, value string) *Request {
	r.Headers[key] = value
	return r
}

// WithHeaders sets multiple request headers.
func (r *Request) WithHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.Headers[k] = v
	}
	return r
}
