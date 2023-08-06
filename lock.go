package redis_lock

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	//go:embed script/lua/unlock.lua
	luaUnlock              string
	ErrFailedToPreemptLock = errors.New("抢锁失败")
	ErrorNotHoldingLock    = errors.New("解锁失败")
)

type Client struct {
	client redis.Cmdable
	valuer func() string
}

type Lock struct {
	client redis.Cmdable
	key    string
	value  string
}

func NewClient(client redis.Cmdable) *Client {
	return &Client{
		client: client,
		valuer: func() string {
			return uuid.New().String()
		},
	}
}

func NewLock(client redis.Cmdable, key string, value string) *Lock {
	return &Lock{
		client: client,
		key:    key,
		value:  value,
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	value := c.valuer()
	ok, err := c.client.SetNX(ctx, key, value, expiration).Result()

	if err != nil {
		return nil, err
	}

	if !ok {
		// 这个key已经被抢了，所以redis设置不上去
		return nil, ErrFailedToPreemptLock
	}
	return NewLock(c.client, key, value), nil
}

func (l *Lock) UnLock(ctx context.Context) error {
	// unlock的逻辑就是：把传入的key值，和当前在redis中的key对应的value进行比对。如果相同，说明unlock的是自己的锁
	// 由于redis的操作都是异步的。所以 check和delete操作应该通过script在redis中一次执行
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()

	if err != nil {
		// 网络错误
		return err
	}

	if res == 0 {
		// 不是自己的key
		// key不存在
		return ErrorNotHoldingLock
	}

	return nil
}
