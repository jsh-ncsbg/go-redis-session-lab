package user

import "context"

type NoopUserCache struct{}

func (n *NoopUserCache) Get(context.Context, int64) (*User, error) {
	return nil, ErrCacheMiss
}

func (n *NoopUserCache) Set(context.Context, *User) error {
	return nil
}

func (n *NoopUserCache) Delete(context.Context, int64) error {
	return nil
}
