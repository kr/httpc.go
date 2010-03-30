package httpc

import (
	"http"
	"net"
	"os"
	"strings"
)

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func dial(addr string) (*http.ClientConn, os.Error) {
	if !hasPort(addr) {
		addr += ":http"
	}
	sock, err := net.Dial("tcp", "", addr)
	if err != nil {
		return nil, err
	}
	return http.NewClientConn(sock, nil), nil
}
