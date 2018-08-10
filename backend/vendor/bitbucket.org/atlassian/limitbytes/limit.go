/*
Package limitbytes implements a middleware to enforce limits on the http request body size.

This is a thin wrapper around `http.MaxBytesReader`. This fails the request early
before reading data if content-length is known, otherwise returns a custom
error type `ErrTooLarge` by parsing the error from `http.MaxBytesReader`. This
also sets the response code to 413 (`http.StatusRequestEntityTooLarge`).

An optional callback can be provided to handle failure case (but this only
works if the content-length is available).
*/
package limitbytes

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// New is a middleware that limits the request body size to n bytes by wrapping
// the request body using http.MaxBytesReader.
// If content-length is available, this fails early without calling the next
// handler.
// An optional callback can be passed in which is called if content-length is
// available and request size is greater than n.
func New(n int64, callback func(http.ResponseWriter, *http.Request, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				if callback != nil {
					callback(w, r, ErrTooLarge{Err: fmt.Errorf("content-length %d is greater than %d", r.ContentLength, n)})
					return
				}
				w.Header().Set("Connection", "close")
				http.Error(w, "Request entity too large.", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = newLimitBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

type limitBytesReader struct {
	w http.ResponseWriter
	r io.ReadCloser
	n int64
}

// newLimitBytesReader returns new instance
func newLimitBytesReader(w http.ResponseWriter, r io.ReadCloser, n int64) *limitBytesReader {
	r = http.MaxBytesReader(w, r, n)
	return &limitBytesReader{w: w, r: r, n: n}
}

func (l *limitBytesReader) Read(p []byte) (n int, err error) {
	var num, e = l.r.Read(p)
	if e != nil {
		// this is not ideal :( but MaxBytesReader does not return a typed error
		if strings.Contains(e.Error(), "http: request body too large") {
			return num, ErrTooLarge{Err: e}
		}
	}
	return num, e
}

func (l *limitBytesReader) Close() error {
	return l.r.Close()
}

// ErrTooLarge error is used to indicate that the request size is too large
type ErrTooLarge struct {
	Err error // underlying error
}

func (r ErrTooLarge) Error() string {
	return fmt.Sprintf("request entity too large error: %s", r.Err.Error())
}
