package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gianglt2198/platforms/internal/core"
)

type HTTPRequest struct {
	req     *http.Request
	payload interface{}
	meta    map[string]interface{}
}

func (r *HTTPRequest) Context() context.Context         { return r.req.Context() }
func (r *HTTPRequest) Protocol() string                 { return "http" }
func (r *HTTPRequest) Method() string                   { return r.req.Method }
func (r *HTTPRequest) Path() string                     { return r.req.URL.Path }
func (r *HTTPRequest) Payload() interface{}             { return r.payload }
func (r *HTTPRequest) Metadata() map[string]interface{} { return r.meta }

type HTTPResponse struct {
	payload interface{}
	err     error
	meta    map[string]interface{}
	status  int
	headers map[string][]string
}

func (r *HTTPResponse) Payload() interface{}             { return r.payload }
func (r *HTTPResponse) Error() error                     { return r.err }
func (r *HTTPResponse) Metadata() map[string]interface{} { return r.meta }

func WrapHTTPHandler(handler http.Handler, middlewares ...core.Middleware) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &HTTPRequest{
			req:  r,
			meta: make(map[string]interface{}),
		}

		var h core.Handler = func(r core.Request) (core.Response, error) {
			// Create response recorder
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req.req)

			return &HTTPResponse{
				payload: rw.Body.Bytes(),
				status:  rw.Code,
				headers: rw.Header(),
				meta:    make(map[string]interface{}),
			}, nil
		}

		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}

		resp, err := h(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write response
		httpResp := resp.(*HTTPResponse)
		for k, v := range httpResp.headers {
			w.Header()[k] = v
		}
		w.WriteHeader(httpResp.status)
		if httpResp.payload != nil {
			if _, err := w.Write(httpResp.payload.([]byte)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}
