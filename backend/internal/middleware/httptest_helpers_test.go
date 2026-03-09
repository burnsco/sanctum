package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
)

func newRequest(target string) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
}
