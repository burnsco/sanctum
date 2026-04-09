package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestMessageHistoryKeyIncludesPaginationAndRotatesOnInvalidate(t *testing.T) {
	oldClient := client
	mr, err := miniredis.Run()
	require.NoError(t, err)

	testClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client = testClient

	t.Cleanup(func() {
		client = oldClient
		require.NoError(t, testClient.Close())
		mr.Close()
	})

	ctx := context.Background()

	firstPage := MessageHistoryKey(ctx, 42, 50, 0)
	nextPage := MessageHistoryKey(ctx, 42, 50, 50)
	smallerPage := MessageHistoryKey(ctx, 42, 25, 0)

	require.NotEqual(t, firstPage, nextPage)
	require.NotEqual(t, firstPage, smallerPage)

	InvalidateRoom(ctx, 42)

	rotatedFirstPage := MessageHistoryKey(ctx, 42, 50, 0)
	require.NotEqual(t, firstPage, rotatedFirstPage)
}
