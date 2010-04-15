package httpc

import (
	"bytes"
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

func echo(c *http.Conn, r *http.Request) {
	c.SetHeader("Content-Type", r.Header["Content-Type"])
	io.Copy(c, r.Body)
}

const port = "12345"

func init() {
	http.Handle("/", HandlerString("hello"))
	http.HandleFunc("/echo", echo)
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
		t.Fatal("nil resp")
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

func TestPost(t *testing.T) {
	ctype, exp := "text/plain", "abc"
	c := NewClient(10, 10)
	if c == nil {
		t.Fatal("nil conn")
	}
	resp, err := Post(c, "http://localhost:"+port+"/echo", ctype, bytes.NewBufferString(exp))
	if err != nil {
		t.Error("unexpedted err", err)
	}
	if resp == nil {
		t.Fatal("nil resp")
	}
	if s := resp.GetHeader("Content-Type"); s != ctype {
		t.Errorf("wrong content type %q", s)
	}
	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("unexpedted err", err)
	}
	if string(got) != exp {
		t.Errorf("expected %q, got %q\n", exp, got)
	}
	//resp.Body.Close()
}

func TestPut(t *testing.T) {
	ctype, exp := "text/plain", "abc"
	c := NewClient(10, 10)
	if c == nil {
		t.Fatal("nil conn")
	}
	resp, err := Put(c, "http://localhost:"+port+"/echo", ctype, bytes.NewBufferString(exp))
	if err != nil {
		t.Error("unexpedted err", err)
	}
	if resp == nil {
		t.Fatal("nil resp")
	}
	if s := resp.GetHeader("Content-Type"); s != ctype {
		t.Errorf("wrong content type %q", s)
	}
	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("unexpedted err", err)
	}
	if string(got) != exp {
		t.Errorf("expected %q, got %q\n", exp, got)
	}
	//resp.Body.Close()
}

