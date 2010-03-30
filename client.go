package httpc

import (
	"container/heap"
	"fmt"
	"http"
	"io"
	"os"
)

// Manages connection pools for all domains.
type client struct {
	LimitGlobal    int
	LimitPerDomain int
	reqs           chan *Request
	poolGetter     chan poolPromise
}

type poolPromise struct {
	name    string
	promise chan *pool
}

func (c *client) getPool(domain string) *pool {
	pp := poolPromise{domain, make(chan *pool)}
	c.poolGetter <- pp
	return <-pp.promise
}

func (c *client) managePools(poolMaker func(string) *pool) {
	pools := make(map[string]*pool)
	for {
		pp := <-c.poolGetter
		p, ok := pools[pp.name]
		if !ok {
			p = poolMaker(pp.name)
			pools[pp.name] = p
		}
		pp.promise <- p
	}
}

func (c *client) accept() {
	for {
		r := <-c.reqs
		d := r.Domain()
		p := c.getPool(d)
		p.reqs <- r
	}
}

func (c *client) drive(incReq, decReq chan *pool) {
	q := new(poolQueue)
	heap.Init(q)
	active := 0
	for {
		var p *pool
		select {
		case p = <-incReq:
			p.pending++
		case p = <-decReq:
			p.active--
			active--
		}

		if p.pos >= 0 {
			heap.Remove(q, p.pos)
		}

		// Don't want pri changing consurrently under our nose, so we copy it for ourselves.
		p.pri = p.wantPri

		if p.pending > 0 && p.active < cap(p.conns) {
			heap.Push(q, p)
		}

		for active < c.LimitGlobal && q.Len() > 0 {
			p = heap.Pop(q).(*pool)
			p.pending--
			p.active++
			active++
			go func() {
				x := (p.execute <- true)
				x = x
			}()
		}
	}
}

func NewClient() *client {
	c := &client{DefaultLimitGlobal, DefaultLimitPerDomain, make(chan *Request), make(chan poolPromise)}
	incReq := make(chan *pool)
	decReq := make(chan *pool)
	go c.managePools(func(addr string) *pool { return newPool(addr, c.LimitPerDomain, incReq, decReq) })
	go c.accept()
	go c.drive(incReq, decReq)
	return c
}

func (c *client) Send(req *Request) (resp *Response, err os.Error) {
	if req.Request.URL, err = http.ParseURL(req.Request.RawURL); err != nil {
		return
	}
	if req.Request.URL.Scheme != "http" {
		return nil, os.ErrorString(fmt.Sprintf("bad scheme %s", req.Request.URL.Scheme))
	}
	c.reqs <- req
	select {
	case resp = <-req.success:
	case err = <-req.failure:
	}
	return
}

func shouldRedirect(status int) bool { return false }

func (c *client) Get(url string) (r *Response, err os.Error) {
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
		if r, err = r.wrap(c.Send(&req)); err != nil {
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

func (c *client) Put(url string, bodyType string, body io.Reader) (r *Response, err os.Error) {
	return
}
