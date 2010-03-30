# httpc.go

An http client library.

This library builds on the http library included with Go, and has all the same
features. In addition, it automates connection pooling, global and per-domain
connection limits, request priorities, caching, etags, and more.

(Note, some of this is not yet implemented.)

The global "connection" limit actually limits pending requests. An idle
connection with no outstanding requests does not count toward this limit.

Because of buggy proxies and servers (especially IIS), this library does not
pipeline requests.

## Example

    resp, err := httpc.Get(nil, "http://example.com/")

## Acknowledgements

Some ideas were derived from [httplib2][]. Soon, some code and tests will be,
too.

[httplib2]: http://code.google.com/p/httplib2/
