package martini

import (
	"context"
	"github.com/go-martini/martini"
	"github.com/slok/go-http-metrics/middleware"
	"net/http"
)

type handler func() martini.Handler

// Handler returns a martini.Handler measuring middleware.
func Handler(handlerID string, m middleware.Middleware) martini.Handler {
	return func(rw http.ResponseWriter, r *http.Request, c martini.Context) {
		reporter := &reporter{
			w: rw.(martini.ResponseWriter),
			r: *r,
		}
		m.Measure(handlerID, reporter, func() {
			c.Next()
		})
	}
}

// responseWriterInterceptor is a simple wrapper to intercept set data on a
// ResponseWriter.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

type reporter struct {
	w martini.ResponseWriter
	r http.Request
}

func (s *reporter) Method() string { return s.r.Method }

func (s *reporter) Context() context.Context { return s.r.Context() }

func (s *reporter) URLPath() string { return s.r.URL.Path }

func (s *reporter) StatusCode() int { return s.w.Status() }

func (s *reporter) BytesWritten() int64 { return int64(s.w.Size()) }
