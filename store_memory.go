package httpc

import (
	"io"
	"os"
)

type entry struct {
	key string
	info map[string]string
	content []byte
	prev, next *entry
}

type memoryStore struct {
	maxBytes, usedBytes int
	entries map[string]*entry
	order entry
}

// Stores responses in a map with bounded size and LRU replacement.
func NewMemoryStore(maxBytes int) Store {
	ms := memoryStore{maxBytes:maxBytes, entries:map[string]*entry{}}
	ms.order.next = &ms.order
	ms.order.prev = &ms.order
	return &ms
}

func (s *memoryStore) Set(key string, info map[string]string, content []byte) {
	if len(content) > s.maxBytes {
		return
	}

	for len(content) > s.maxBytes - s.usedBytes {
		s.Delete(s.order.prev.key)
	}

	e := &entry{key, info, content, &s.order, s.order.next}
	e.next.prev = e
	s.order.next = e
	s.entries[key] = e
	s.usedBytes += len(content)
}

func (s memoryStore) Get(key string) (map[string]string, io.ReadCloser) {
	entry, ok := s.entries[key]
	if !ok {
		return nil, nil
	}
	return entry.info, &stringReadCloser{entry.content, 0}
}

func (s *memoryStore) Delete(key string) {
	e, ok := s.entries[key]
	if !ok {
		return
	}
	s.usedBytes -= len(e.content)
	s.entries[key] = nil, false
	e.prev.next = e.next
	e.next.prev = e.prev
}

type stringReadCloser struct {
	buf []byte
	off int
}

func (x *stringReadCloser) Read(buf []byte) (n int, err os.Error) {
	if x.off >= len(x.buf) {
		x.Close()
		return 0, os.EOF
	}

	n = copy(buf, x.buf[x.off:])
	x.off += n
	if len(buf) > n {
		err = os.EOF
	}
	return
}

func (x *stringReadCloser) Close() os.Error {
	x.buf = []byte{}
	x.off = 0
	return nil
}

