//go:build e2e

package redis_lock

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// locked
// network error
// ErrorFailedToPreemptLock

func TestClient_TryLock_e2e(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	c := NewClient(rdb)
	c.Wait()

	testCases := []struct {
		name string

		key        string
		expiration time.Duration

		before func()
		after  func()

		wantErr  error
		wantLock *Lock
	}{
		{
			name:       "locked",
			key:        "locked-key",
			expiration: time.Minute,
			before:     func() {},
			after: func() {
				res, err := rdb.Del(context.Background(), "locked-key").Uint64()
				require.NoError(t, err)
				require.Equal(t, uint64(1), res)
			},
			wantLock: &Lock{
				key: "locked-key",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()

			l, err := c.TryLock(context.Background(), tc.key, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			tc.after()
			assert.NotNil(t, l.client)
			assert.Equal(t, tc.wantLock.key, l.key)
			assert.NotEmpty(t, l.value)
		})
	}
}

func (c *Client) Wait() {
	for c.client.Ping(context.Background()).Err() != nil {
	}
	return
}
