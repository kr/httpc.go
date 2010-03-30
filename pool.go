package httpc

import (
	"container/heap"
	"container/vector"
	"http"
	"io"
	"os"
)

// A connection pool for one domain+port.
type pool struct {
	addr    string
	reqs    chan *Request
	execute chan bool
	wantPri int
	conns   chan *http.ClientConn

	// managed by client driver
	pri     int
	pos     int
	pending int
	active  int
}

func (p *pool) exec(r *Request) (resp *Response, err os.Error) {
	conns := p.conns
	for {
		conn := <-conns
		if conn == nil {
			conn, err = dial(p.addr)
			if err != nil {
				conns <- nil
				return
			}
		}

		err = conn.Write(&r.Request)
		if err != nil {
			conn.Close()
			conns <- nil
			if perr, ok := err.(*http.ProtocolError); ok && perr == http.ErrPersistEOF {
				continue
			} else if err == io.ErrUnexpectedEOF {
				continue
			} else if err != nil {
				return
			}
		}

		x, err := conn.Read()
		if err != nil {
			conn.Close()
			conns <- nil
			if perr, ok := err.(*http.ProtocolError); ok && perr == http.ErrPersistEOF {
				continue
			} else if err == io.ErrUnexpectedEOF {
				continue
			} else if err != nil {
				return
			}
		}

		done := make(chan bool)
		resp = &Response{Response: x, Body: body{x.Body, done}}

		// When the user is done reading the response, put this conn back into the pool.
		go func() {
			<-done
			conns <- conn
		}()

		return
	}
	panic("can not happen")
}

func (p *pool) hookup(r *Request, decReq chan<- *pool) {
	defer func() { decReq <- p }()

	resp, err := p.exec(r)
	if err != nil {
		r.failure <- err
		return
	}
	r.success <- resp
}

func (p *pool) accept(incReq, decReq chan<- *pool) {
	q := new(requestQueue)
	heap.Init(q)
	for {
		select {
		case r := <-p.reqs:
			heap.Push(q, r)
			p.wantPri = q.At(0).(*Request).Pri
			incReq <- p
		case <-p.execute:
			r := heap.Pop(q).(*Request)
			go p.hookup(r, decReq)
		}
	}
}

func newPool(addr string, limit int, incReq, decReq chan<- *pool) *pool {
	p := &pool{
		addr:    addr,
		pos:     -1,
		reqs:    make(chan *Request),
		execute: make(chan bool),
	}

	p.setLimit(limit)

	go p.accept(incReq, decReq)

	return p
}

func (p *pool) setLimit(limit int) {
	conns := make(chan *http.ClientConn, limit)
	for i := 0; i < limit; i++ {
		conns <- nil
	}
	p.conns = conns
}

func (p *pool) Ready() bool { return false }

type poolQueue struct {
	vector.Vector
}

func (q *poolQueue) Less(i, j int) bool { return i < j }

func (q *poolQueue) Push(x interface{}) {
	pos := q.Len()
	q.Vector.Push(x)
	x.(*pool).pos = pos
}

func (q *poolQueue) Pop() (x interface{}) {
	x = q.Vector.Pop()
	x.(*pool).pos = -1
	return
}

func (q *poolQueue) Swap(i, j int) {
	q.Vector.Swap(i, j)
	q.Vector.At(i).(*pool).pos, q.Vector.At(j).(*pool).pos = i, j
}
