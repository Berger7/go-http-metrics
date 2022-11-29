package martini_test

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	mmetrics "github.com/slok/go-http-metrics/internal/mocks/metrics"
	"github.com/slok/go-http-metrics/metrics"
	"github.com/slok/go-http-metrics/middleware"
	marinimiddleware "github.com/slok/go-http-metrics/middleware/martini"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	tests := map[string]struct {
		handlerID   string
		config      middleware.Config
		req         func() *http.Request
		mock        func(m *mmetrics.Recorder)
		handler     func() martini.Handler
		expRespCode int
		expRespBody string
	}{
		"A default HTTP middleware should call the recorder to measure.": {
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/test", nil)
			},
			mock: func(m *mmetrics.Recorder) {
				expHTTPReqProps := metrics.HTTPReqProperties{
					ID:      "/test",
					Service: "",
					Method:  "POST",
					Code:    "202",
				}
				m.On("ObserveHTTPRequestDuration", mock.Anything, expHTTPReqProps, mock.Anything).Once()
				m.On("ObserveHTTPResponseSize", mock.Anything, expHTTPReqProps, int64(5)).Once()

				expHTTPProps := metrics.HTTPProperties{
					ID:      "/test",
					Service: "",
				}
				m.On("AddInflightRequests", mock.Anything, expHTTPProps, 1).Once()
				m.On("AddInflightRequests", mock.Anything, expHTTPProps, -1).Once()
			},
			handler: func() martini.Handler {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(202)
					_, _ = w.Write([]byte("test1"))
				}
			},
			expRespCode: 202,
			expRespBody: "test1",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testAssert := assert.New(t)

			// Setup
			mr := &mmetrics.Recorder{}
			test.mock(mr)
			req := test.req()

			mw := middleware.New(middleware.Config{Recorder: mr})
			m := newMartini()
			m.AddRoute(req.Method, req.URL.Path, marinimiddleware.Handler(test.handlerID, mw), test.handler())

			// Execute
			resp := httptest.NewRecorder()
			m.ServeHTTP(resp, req)

			// Assert
			mr.AssertExpectations(t)
			testAssert.Equal(test.expRespCode, resp.Result().StatusCode)
			testAssert.Equal(test.expRespBody, resp.Body.String())
		})
	}
}

func newMartini() *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()

	m.Use(martini.Recovery())
	m.Use(render.Renderer())

	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{Martini: m, Router: r}
}
