package httpc

import (
	"container/vector"
	"http"
	"os"
)

type Request struct {
	http.Request
	Pri int

	success chan *Response
	failure chan os.Error
}

func (r *Request) Domain() string {
	if r.Request.URL == nil {
		return ""
	}
	return r.Request.URL.Host
}

type requestQueue struct {
	vector.Vector
}

func (q requestQueue) Less(i, j int) bool { return i < j }
