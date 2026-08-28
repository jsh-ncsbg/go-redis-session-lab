package session

import "context"

type NoopSessionCache struct{}

func (n *NoopSessionCache) Get(context.Context, string) (*Session, error) {
	return nil, ErrCacheMiss
}

func (n *NoopSessionCache) Set(context.Context, *Session) error {
	return nil
}

func (n *NoopSessionCache) Delete(context.Context, string) error {
	return nil
}
