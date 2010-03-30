// An http client library.
//
// This library builds on the http library included with Go, and has all the
// same features. In addition, it automates connection pooling, global and
// per-domain connection limits, request priorities, caching, etags, and more.
//
// Because of buggy proxies and servers (especially IIS), this library does not
// pipeline requests.
//
// For example,
//
//   resp, err := httpc.Get("http://example.com/")
package httpc

import (
	"io"
	"os"
)

const (
	DefaultLimitGlobal    = 40
	DefaultLimitPerDomain = 6
)

const DefaultPri = 5000

var DefaultClient = NewClient()

// Shorthand for DefaultClient.Get(url)
func Get(url string) (r *Response, err os.Error) {
	return DefaultClient.Get(url)
}

func PostParams(url string, params map[string]string) (r *Response, err os.Error) {
	return
}

func Put(url string, bodyType string, body io.Reader) (r *Response, err os.Error) {
	return
}

func Delete(url string) (r *Response, err os.Error) {
	return
}
