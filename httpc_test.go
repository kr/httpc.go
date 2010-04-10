package httpc

import (
	"http"
	"io"
	"io/ioutil"
	"net"
	"testing"
)

type HandlerString string

func (s HandlerString) ServeHTTP(c *http.Conn, r *http.Request) {
	io.WriteString(c, string(s))
}

const port = "12345"

func init() {
	http.Handle("/", HandlerString("hello"))
	go func() {
		l, e := net.Listen("tcp", ":"+port)
		if e != nil {
			panic(e)
		}
		e = http.Serve(l, nil)
		l.Close()
		panic(e)
	}()
}

func TestGet(t *testing.T) {
	c := NewClient(10, 10)
	if c == nil {
		t.Fatal("nil conn")
	}
	resp, err := Get(c, "http://localhost:"+port+"/")
	if err != nil {
		t.Error("unexpedted err", err)
	}
	if resp == nil {
		t.Error("nil resp")
	}
	s, err := ioutil.ReadAll(resp[0].Body)
	if err != nil {
		t.Error("unexpedted err", err)
	}
	if string(s) != "hello" {
		t.Errorf("expected hello, got %q\n", s)
	}
	//resp.Body.Close()
}
