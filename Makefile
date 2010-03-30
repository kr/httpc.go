include $(GOROOT)/src/Make.$(GOARCH)

TARG=httpc
GOFILES=\
	httpc.go\
	client.go\
	conn.go\
	pool.go\
	request.go\
	response.go\

include $(GOROOT)/src/Make.pkg
