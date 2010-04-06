package httpc

import (
	"container/vector"
	"http"
	"os"
)

type Request struct {
	http.Request
	Pri int

	success chan *http.Response
	failure chan os.Error
}

type requestQueue struct {
	vector.Vector
}

func (q requestQueue) Less(i, j int) bool { return i < j }
