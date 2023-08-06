package redis_lock

import (
	"context"
	_ "embed"
	"errors"
	"redis-lock/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// locked
// network error
// ErrorFailedToPreemptLock

func TestClient_TryLock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name string

		key        string
		expiration time.Duration

		mock func() redis.Cmdable

		wantErr  error
		wantLock *Lock
	}{
		{
			name:       "locked",
			key:        "locked-key",
			expiration: time.Minute,
			mock: func() redis.Cmdable {
				rdb := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				rdb.EXPECT().
					SetNX(gomock.Any(), "locked-key", gomock.Any(), time.Minute).
					Return(res)
				return rdb
			},
			wantLock: &Lock{
				key: "locked-key",
			},
		},
		{
			name:       "network error",
			key:        "network-key",
			expiration: time.Minute,
			mock: func() redis.Cmdable {
				rdb := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, errors.New("network error"))
				rdb.EXPECT().
					SetNX(gomock.Any(), "network-key", gomock.Any(), time.Minute).
					Return(res)
				return rdb
			},
			wantErr: errors.New("network error"),
		},

		{
			name:       "failed to lock",
			key:        "failed-key",
			expiration: time.Minute,
			mock: func() redis.Cmdable {
				rdb := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				rdb.EXPECT().
					SetNX(gomock.Any(), "failed-key", gomock.Any(), time.Minute).
					Return(res)
				return rdb
			},
			wantErr: ErrFailedToPreemptLock,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(tc.mock())
			l, err := c.TryLock(context.Background(), tc.key, tc.expiration)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.NotNil(t, l.client)
			assert.Equal(t, tc.wantLock.key, l.key)
			assert.NotEmpty(t, l.value)
		})
	}
}
