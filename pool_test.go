package httpc

import (
	"http"
	"testing"
)

func TestRequestQueueLess(t *testing.T) {
	q := new(requestQueue)
	var a, b, c, d http.Request

	a.Header = map[string]string {
		"X-Pri": "1",
	}
	b.Header = map[string]string {
		"X-Pri": "2",
	}
	c.Header = map[string]string {
		"X-Pri": "1",
	}
	d.Header = map[string]string {} // default X-Pri: 5000

	q.Push(&clientRequest{r:&a})
	q.Push(&clientRequest{r:&b})
	q.Push(&clientRequest{r:&c})
	q.Push(&clientRequest{r:&d})

	if !q.Less(0, 1) {
		t.Error("want a < b")
	}
	if q.Less(1, 0) {
		t.Error("want b > a")
	}

	if !q.Less(2, 1) {
		t.Error("want c < b")
	}
	if q.Less(1, 2) {
		t.Error("want b > c")
	}

	if q.Less(0, 2) {
		t.Error("want a == c")
	}
	if q.Less(2, 0) {
		t.Error("want c == a")
	}

	if !q.Less(0, 3) {
		t.Error("want a < d")
	}
	if q.Less(3, 0) {
		t.Error("want d > a")
	}

}
