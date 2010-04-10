// An http client library.
//
// This library builds on the http library included with Go, and has all the
// same features. In addition, it automates connection pooling, global and
// per-domain connection limits, request priorities, caching, etags, and more.
//
// The global "connection" limit actually limits pending requests. An idle
// connection with no outstanding requests does not count toward this limit.
//
// Because of buggy proxies and servers (especially IIS), this library does not
// pipeline requests.
//
// A simple example:
//
//   resp, err := httpc.Get(nil, "http://example.com/")
package httpc

import (
	"fmt"
	"http"
	"io"
	"os"
	"strconv"
)

// This interface is for sending HTTP requests.
type Sender interface {
	Send(*http.Request) (*http.Response, os.Error)
}

const (
	DefaultLimitGlobal    = 40
	DefaultLimitPerDomain = 6
)

// Used for requests made by Get, Put, Post, PostParams, and Delete.
const DefaultPri = 5000

const DefaultMemoryStoreSize = 50000000 // 50MB

var (
	DefaultClient = NewClient(DefaultLimitGlobal, DefaultLimitPerDomain)
	DefaultStore = NewMemoryStore(DefaultMemoryStoreSize)
	DefaultSender = NewCache(DefaultStore, DefaultClient)
)

func prepend(r *http.Response, rs []*http.Response) []*http.Response {
	nrs := make([]*http.Response, len(rs) + 1)
	nrs[0] = r
	copy(nrs[1:], rs)
	return nrs
}

func getHeader(r *http.Request, key string) (value string) {
    return r.Header[http.CanonicalHeaderKey(key)]
}

func Send(s Sender, req *http.Request) (resp *http.Response, err os.Error) {
	if s == nil {
		s = DefaultSender
	}
	header := req.Header
	req.Header = map[string]string{}
	for k, v := range header {
		req.Header[http.CanonicalHeaderKey(k)] = v
	}
	return s.Send(req)
}

// Much like http.Get. If s is nil, uses DefaultSender.
func Get(s Sender, url string) (rs []*http.Response, err os.Error) {
	for redirect := 0; ; redirect++ {
		if redirect >= 10 {
			err = os.ErrorString("stopped after 10 redirects")
			break
		}

		var req http.Request
		req.RawURL = url
		req.Header = map[string]string{}
		req.Header["X-Pri"] = strconv.Itoa(DefaultPri)
		r, err := Send(s, &req)
		if err != nil {
			break
		}
		rs = prepend(r, rs)
		if shouldRedirect(r.StatusCode) {
			r.Body.Close()
			if url = r.GetHeader("Location"); url == "" {
				err = os.ErrorString(fmt.Sprintf("%d response missing Location header", r.StatusCode))
				break
			}
			continue
		}

		return
	}

	err = &http.URLError{"Get", url, err}
	return
}

func Post(s Sender, url string, bodyType string, body io.Reader) (r *http.Response, err os.Error) {
	return
}

func PostParams(s Sender, url string, params map[string]string) (r *http.Response, err os.Error) {
	return
}

func Put(s Sender, url string, bodyType string, body io.Reader) (r *http.Response, err os.Error) {
	return
}

func Delete(s Sender, url string) (r *http.Response, err os.Error) {
	return
}
