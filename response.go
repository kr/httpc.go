package httpc

import (
	"http"
	"io"
	"os"
)

type Response struct {
	*http.Response
	Prev *Response
	Body io.ReadCloser
}

type body struct {
	rc   io.ReadCloser
	done chan bool
}

func (b body) Read(p []byte) (n int, err os.Error) {
	n, err = b.rc.Read(p)
	if err == os.EOF {
		close(b.done)
	}
	return
}

func (b body) Close() os.Error {
	defer close(b.done)
	return b.rc.Close()
}

func (r *Response) wrap(wrapper *Response, err os.Error) (*Response, os.Error) {
	if wrapper != nil {
		wrapper.Prev = r
	}
	return wrapper, err
}
