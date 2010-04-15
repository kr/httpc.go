package httpc

import (
	"bytes"
	"http"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	fresh = iota
	stale
)

func normURL(url string) string {
	u, err := http.ParseURL(url)
	if err != nil {
		return url // This is a cheap hack
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if u.Path == "" {
		u.Path = "/"
	}
	return u.String()
}

// When implementing this interface, let errors silently fail, as if the entry
// in question never existed.
type Store interface {
	Set(key string, info map[string]string, content []byte)
	Get(key string) (map[string]string, io.ReadCloser)
	Delete(key string)
}

type cache struct {
	store Store
    next Sender
}

func NewCache(store Store, next Sender) Sender {
	return cache{store, next}
}

var ignoreHeaders = map[string]bool {
	"Content-Encoding": true,
	"Transfer-Encoding": true,
}

func (c cache) updateStore(key string, body []byte, req *http.Request, resp *http.Response) {
	if key == "" {
		return
	}
	info := map[string]string{}
	for k, v := range resp.Header {
		if _, ok := ignoreHeaders[k]; !ok {
			info[k] = v
		}
	}

	// TODO deal with Vary header

	info["Status"] = resp.Status
	c.store.Set(key, info, body)
}

func (c cache) sendAndUpdate(req *http.Request, key string) (*http.Response, os.Error) {
	resp, err := c.next.Send(req)
	if err != nil {
		return resp, err
	}
	t := &tee{resp.Body, bytes.NewBuffer([]byte{}), false, func(t *tee) {
		c.updateStore(key, t.buf.Bytes(), req, resp)
	}}
	resp.Body = t
	return resp, err
}

func valueOrDefault(value, def string) string {
	if value == "" {
		return def
	}
	return value
}

func (c cache) Send(req *http.Request) (resp *http.Response, err os.Error) {
	method := valueOrDefault(req.Method, "GET")
	key := normURL(req.RawURL)
	info, content := c.store.Get(key)
	if info != nil {
		state := state(info, req.Header)
		if state == fresh {
			response := cacheResponse(info)
			response.Body = content
			response.AddHeader("Via", "1.1 internal (httpc.go)")
			response.Status = info["Status"]
			statusCode, ok := info["StatusCode"]
			if !ok {
				statusCode = "200"
			}
			response.StatusCode, err = strconv.Atoi(statusCode)
			if err != nil {
				err = nil
				response.StatusCode = 200
			}
			return response, nil
		}

		if state == stale {
			// TODO modify request headers
			return nil, os.NewError("not implemented")
		}

		return nil, os.NewError("stub")
		resp, err = c.sendAndUpdate(req, key)
		if err != nil {
			return
		}

		if resp.StatusCode == 304 && method == "GET" {
			return nil, os.NewError("not implemented")
		} else if resp.StatusCode == 200 {
		} else {
			return nil, os.NewError("not implemented")
		}
	} else {
		resp, err = c.sendAndUpdate(req, key)
	}

	return
}

func cacheResponse(info map[string]string) *http.Response {
	hr := &http.Response{}
	hr.Header = info
	return hr
}

func state(info, header map[string]string) int {
	return fresh
}

type tee struct {
	rc io.ReadCloser
	buf *bytes.Buffer
	isDone bool
	onDone func(*tee)
}

func (t *tee) Read(buf []byte) (n int, err os.Error) {
	n, err = t.rc.Read(buf)
	t.buf.Write(buf[0:n])
	if err == os.EOF && !t.isDone {
		t.isDone = true
		t.onDone(t)
	}
	return
}

func (t *tee) Close() os.Error {
	_, err := t.buf.ReadFrom(t.rc)
	if err == nil && !t.isDone {
		t.isDone = true
		t.onDone(t)
	}

	return t.rc.Close()
}

