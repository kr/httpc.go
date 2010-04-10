package httpc

import (
	"container/heap"
	"fmt"
	"http"
	"os"
)

// Manages connection pools for all domains.
type client struct {
	limitGlobal    int
	limitPerDomain int
	reqs           chan *clientRequest
	poolGetter     chan poolPromise
}

type clientRequest struct {
	r *http.Request
	success chan *http.Response
	failure chan os.Error
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
		d := r.r.URL.Host
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

		for active < c.limitGlobal && q.Len() > 0 {
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

func NewClient(limitGlobal, limitPerDomain int) Sender {
	c := &client{limitGlobal, limitPerDomain, make(chan *clientRequest), make(chan poolPromise)}
	incReq := make(chan *pool)
	decReq := make(chan *pool)
	go c.managePools(func(addr string) *pool { return newPool(addr, c.limitPerDomain, incReq, decReq) })
	go c.accept()
	go c.drive(incReq, decReq)
	return c
}

func (c *client) Send(req *http.Request) (resp *http.Response, err os.Error) {
	if req.URL, err = http.ParseURL(req.RawURL); err != nil {
		return
	}
	if req.URL.Scheme != "http" {
		return nil, os.ErrorString(fmt.Sprintf("bad scheme %s", req.URL.Scheme))
	}
	cr := &clientRequest{req, make(chan *http.Response), make(chan os.Error)}
	c.reqs <- cr
	select {
	case resp = <-cr.success:
	case err = <-cr.failure:
	}
	return
}

func shouldRedirect(status int) bool { return false }
