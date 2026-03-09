package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
)

func newRequest(method, target string, body io.Reader) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), method, target, body)
}
