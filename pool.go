package httpc

import (
	"container/heap"
	"container/vector"
	"http"
	"io"
	"os"
	"strconv"
)

// A connection pool for one domain+port.
type pool struct {
	addr    string
	reqs    chan *clientRequest
	execute chan bool
	wantPri int
	conns   chan *http.ClientConn

	// managed by client driver
	pri     int
	pos     int
	pending int
	active  int
}

func (p *pool) exec(r *http.Request) (resp *http.Response, err os.Error) {
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

		err = conn.Write(r)
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
		x.Body = body{x.Body, done}
		resp = x

		// When the user is done reading the response, put this conn back into the pool.
		go func() {
			<-done
			conns <- conn
		}()

		return
	}
	panic("can not happen")
}

func (p *pool) hookup(cr *clientRequest, decReq chan<- *pool) {
	defer func() { decReq <- p }()

	resp, err := p.exec(cr.r)
	if err != nil {
		cr.failure <- err
		return
	}
	cr.success <- resp
}

func (p *pool) accept(incReq, decReq chan<- *pool) {
	q := new(requestQueue)
	heap.Init(q)
	for {
		select {
		case r := <-p.reqs:
			heap.Push(q, r)
			pri, err := strconv.Atoi(getHeader(q.At(0).(*clientRequest).r, "X-Pri"))
			if err != nil {
				pri = DefaultPri
			}
			p.wantPri = pri
			incReq <- p
		case <-p.execute:
			cr := heap.Pop(q).(*clientRequest)
			go p.hookup(cr, decReq)
		}
	}
}

func newPool(addr string, limit int, incReq, decReq chan<- *pool) *pool {
	p := &pool{
		addr:    addr,
		pos:     -1,
		reqs:    make(chan *clientRequest),
		execute: make(chan bool),
	}

	p.conns = make(chan *http.ClientConn, limit)
	for i := 0; i < limit; i++ {
		p.conns <- nil
	}

	go p.accept(incReq, decReq)

	return p
}

func (p *pool) Ready() bool { return false }

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

type requestQueue struct {
	vector.Vector
}

func (q requestQueue) Less(i, j int) bool { return i < j }

