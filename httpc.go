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
	"fmt"
	"http"
	"io"
	"os"
)

// This interface is for sending HTTP requests.
type Sender interface {
	Send(*Request) (*Response, os.Error)
}

const (
	DefaultLimitGlobal    = 40
	DefaultLimitPerDomain = 6
)

const DefaultPri = 5000

var DefaultClient = NewClient()

func Get(s Sender, url string) (r *Response, err os.Error) {
	if s == nil {
		s = DefaultClient
	}

	for redirect := 0; ; redirect++ {
		if redirect >= 10 {
			err = os.ErrorString("stopped after 10 redirects")
			break
		}

		var req Request
		req.success = make(chan *Response)
		req.failure = make(chan os.Error)
		req.Request.RawURL = url
		req.Pri = DefaultPri
		if r, err = r.wrap(s.Send(&req)); err != nil {
			break
		}
		if shouldRedirect(r.Response.StatusCode) {
			r.Response.Body.Close()
			if url = r.GetHeader("Location"); url == "" {
				err = os.ErrorString(fmt.Sprintf("%d response missing Location header", r.Response.StatusCode))
				break
			}
			continue
		}

		return
	}

	err = &http.URLError{"Get", url, err}
	return
}

func Post(s Sender, url string, bodyType string, body io.Reader) (r *Response, err os.Error) {
	return
}

func PostParams(s Sender, url string, params map[string]string) (r *Response, err os.Error) {
	return
}

func Put(s Sender, url string, bodyType string, body io.Reader) (r *Response, err os.Error) {
	return
}

func Delete(s Sender, url string) (r *Response, err os.Error) {
	return
}
