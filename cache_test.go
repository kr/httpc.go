package httpc

import (
	"http"
	"os"
	"testing"
)

type T testing.T

func (t *T) want(b bool, format string, args ...interface{}) {
	tt := (*testing.T)(t)
	if !b {
			tt.Errorf("want " + format, args)
	}
}

func (t *T) need(b bool, format string, args ...interface{}) {
	tt := (*testing.T)(t)
	if !b {
			tt.Fatalf("need " + format, args)
	}
}

func (t *T) assertEQ(a, b interface{}, desc string) {
	t.want(a == b, "%s %#v == %#v", desc, a, b)
}

func (t *T) assertNE(a, b interface{}) {
	t.want(a != b, "%#v != %#v", a, b)
}

func (t *T) noErr(err os.Error) {
	t.need(err == nil, "nil err, got %#v", err)
}

func TestUrlNorm(tt *testing.T) {
 t := (*T)(tt)
	t.assertEQ("http://example.org/", normURL("http://example.org"), "")
	t.assertEQ("http://example.org/", normURL("http://EXAMple.org"), "")
	t.assertEQ("http://example.org/?=b", normURL("http://EXAMple.org?=b"), "")
	t.assertEQ("http://example.org/mypath?a=b", normURL("http://EXAMple.org/mypath?a=b"), "")
	t.assertEQ("http://localhost:80/", normURL("http://localhost:80"), "")
	t.assertEQ("http://localhost:80/", normURL("HTTP://LOCALHOST:80"), "")
	t.assertEQ("/", normURL("/"), "")
	t.assertEQ(normURL("http://www"), normURL("http://WWW"), "")
}

var testEtags = make(map[string]*http.Response)

var dummyResponses = map[string]*http.Response {
	"http://localhost/304/test_etag.txt": &http.Response{
		Header: map[string]string{
			"ETag": "abc",
		},
		Body: &stringReadCloser{[]byte("dummy contents"), 0},
	},
}

func init() {
	for _, r := range dummyResponses {
		if etag, ok := r.Header["ETag"]; ok {
			testEtags[etag] = r
		}
	}
}

type dummySender struct {
}

func (ds *dummySender) Send(req *http.Request) (*http.Response, os.Error) {
	if req.Header == nil {
		req.Header = map[string]string{}
	}
	if etag, ok := req.Header["If-None-Match"]; ok {
		if resp, ok := testEtags[etag]; ok {
			resp.Status = "304 Not Modified"
			resp.StatusCode = 304
			return resp, nil
		}
	}
	resp, ok := dummyResponses[req.RawURL]
	if !ok {
		return nil, os.NewError("no response")
	}
	resp.Status = "200 OK"
	resp.StatusCode = 200
	return resp, nil
}

func TestGetOnlyIfCachedCacheHit(tt *testing.T) {
	t := (*T)(tt)

	// Test that we can do a GET with cache and 'only-if-cached'
	mc := NewMemoryStore(10000000)
	s := NewCache(mc, &dummySender{})
	url := "http://localhost/304/test_etag.txt"
	resp, err := s.Send(&http.Request{RawURL:url}) // Put it in the cache
	t.noErr(err)
	if resp == nil {
		tt.Fatal("got nil resp")
	}
	t.assertEQ(resp.GetHeader("Via"), "", "Via")
	resp.Body.Close()
	resp, err = s.Send(&http.Request{RawURL:url, Header:map[string]string{"Cache-Control":"Only-If-Cached"}})
	t.noErr(err)
	if resp == nil {
		tt.Fatal("got nil resp")
	}
	t.assertEQ(resp.GetHeader("Via"), "1.1 internal (httpc.go)", "Via")
	t.assertEQ(resp.StatusCode, 200, "status")
}

