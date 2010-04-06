include $(GOROOT)/src/Make.$(GOARCH)

TARG=httpc
GOFILES=\
	httpc.go\
	cache.go\
	client.go\
	conn.go\
	pool.go\
	store_file.go\
	store_memory.go\

include $(GOROOT)/src/Make.pkg
