package cache

import (
	"strings"
	"testing"
)

func TestInitRedisReturnsErrorForInvalidURL(t *testing.T) {
	t.Cleanup(func() { client = nil })

	err := InitRedis("redis://%zz")
	if err == nil {
		t.Fatal("expected invalid Redis URL to return an error")
	}
	if !strings.Contains(err.Error(), "invalid REDIS_URL") {
		t.Fatalf("expected invalid REDIS_URL error, got %v", err)
	}
	if GetClient() != nil {
		t.Fatal("expected Redis client to be nil after failed init")
	}
}
