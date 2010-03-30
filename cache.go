package httpc

// This will eventually be a cache that stores responses on disk. It is currently a stub.
func NewFileCache(path string, maxBytes int, next Sender) Sender {
	return next
}

// Stores responses in a map. Currently just a stub.
func NewMemCache(maxBytes int, next Sender) Sender {
	return next
}
