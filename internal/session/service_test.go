package session

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

type serviceCall struct {
	method, token string
	session       *Session
}

type callRecorder []serviceCall

func (r *callRecorder) record(call serviceCall) {
	*r = append(*r, call)
}

type cacheSpy struct {
	calls                     *callRecorder
	session                   *Session
	getErr, setErr, deleteErr error
}

func (s *cacheSpy) Get(_ context.Context, token string) (*Session, error) {
	s.calls.record(serviceCall{method: "cache.Get", token: token})
	return s.session, s.getErr
}

func (s *cacheSpy) Set(_ context.Context, session *Session) error {
	s.calls.record(serviceCall{method: "cache.Set", session: session})
	return s.setErr
}

func (s *cacheSpy) Delete(_ context.Context, token string) error {
	s.calls.record(serviceCall{method: "cache.Delete", token: token})
	return s.deleteErr
}

type storeSpy struct {
	calls                        *callRecorder
	session                      *Session
	getErr, createErr, deleteErr error
}

func (s *storeSpy) Get(_ context.Context, token string) (*Session, error) {
	s.calls.record(serviceCall{method: "store.Get", token: token})
	return s.session, s.getErr
}

func (s *storeSpy) Create(_ context.Context, session *Session) error {
	s.calls.record(serviceCall{method: "store.Create", session: session})
	return s.createErr
}

func (s *storeSpy) Delete(_ context.Context, token string) error {
	s.calls.record(serviceCall{method: "store.Delete", token: token})
	return s.deleteErr
}

func assertCalls(t *testing.T, got, want callRecorder) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Fatalf("calls = %v, want %v", got, want)
	}
}

func TestSessionServiceGet(t *testing.T) {
	ctx := context.Background()
	session := &Session{TokenHash: "token-hash", ExpiresAt: time.Now().Add(time.Hour)}

	t.Run("cache hit skips store", func(t *testing.T) {
		var calls callRecorder
		cache := &cacheSpy{calls: &calls, session: session}
		store := &storeSpy{calls: &calls}
		got, err := (&SessionService{cache: cache, repo: store}).Get(ctx, session.TokenHash)

		if err != nil || got != session {
			t.Fatalf("Get() = (%p, %v), want (%p, nil)", got, err, session)
		}
		assertCalls(t, calls, callRecorder{
			{method: "cache.Get", token: session.TokenHash},
		})
	})

	t.Run("cache hit ignores expired session", func(t *testing.T) {
		var calls callRecorder
		expiredSession := &Session{TokenHash: "token-hash", ExpiresAt: time.Now().Add(-time.Second)}

		cache := &cacheSpy{calls: &calls, session: expiredSession}
		store := &storeSpy{calls: &calls}
		got, err := (&SessionService{cache: cache, repo: store}).Get(ctx, expiredSession.TokenHash)

		if !errors.Is(err, ErrSessionNotFound) {
			t.Fatalf("Get() = (%p, %v), want (nil, %v)", got, err, ErrSessionNotFound)
		}
		assertCalls(t, calls, callRecorder{
			{method: "cache.Get", token: session.TokenHash},
			{method: "cache.Delete", token: session.TokenHash},
		})
	})

	for _, tt := range []struct {
		name   string
		getErr error
		setErr error
	}{
		{name: "cache miss", getErr: ErrCacheMiss},
		{name: "cache failure", getErr: errors.New("cache get failed"), setErr: errors.New("cache set failed")},
	} {
		t.Run(tt.name+" falls back and warms cache", func(t *testing.T) {
			var calls callRecorder
			cache := &cacheSpy{calls: &calls, getErr: tt.getErr, setErr: tt.setErr}
			store := &storeSpy{calls: &calls, session: session}
			got, err := (&SessionService{cache: cache, repo: store}).Get(ctx, session.TokenHash)

			if err != nil || got != session {
				t.Fatalf("Get() = (%p, %v), want (%p, nil)", got, err, session)
			}
			assertCalls(t, calls, callRecorder{
				{method: "cache.Get", token: session.TokenHash},
				{method: "store.Get", token: session.TokenHash},
				{method: "cache.Set", session: session},
			})
		})
	}

	t.Run("store error is returned", func(t *testing.T) {
		var calls callRecorder
		wantErr := errors.New("store failed")
		cache := &cacheSpy{calls: &calls, getErr: ErrCacheMiss}
		store := &storeSpy{calls: &calls, getErr: wantErr}
		got, err := (&SessionService{cache: cache, repo: store}).Get(ctx, session.TokenHash)

		if got != nil || !errors.Is(err, wantErr) {
			t.Fatalf("Get() = (%v, %v), want (nil, %v)", got, err, wantErr)
		}
		assertCalls(t, calls, callRecorder{
			{method: "cache.Get", token: session.TokenHash},
			{method: "store.Get", token: session.TokenHash},
		})
	})
}

func TestSessionServiceCreate(t *testing.T) {
	ctx := context.Background()
	session := &Session{TokenHash: "token-hash"}

	for _, tt := range []struct {
		name     string
		storeErr error
	}{
		{name: "success"},
		{name: "store failure stops", storeErr: errors.New("store failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls callRecorder
			store := &storeSpy{calls: &calls, createErr: tt.storeErr}
			err := (&SessionService{repo: store}).Create(ctx, session)

			if !errors.Is(err, tt.storeErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.storeErr)
			}
			wantCalls := callRecorder{{method: "store.Create", session: session}}
			assertCalls(t, calls, wantCalls)
		})
	}
}

func TestSessionServiceRevoke(t *testing.T) {
	ctx := context.Background()

	for _, tt := range []struct {
		name     string
		storeErr error
		cacheErr error
	}{
		{name: "cache error is returned", cacheErr: errors.New("cache failed")},
		{name: "store failure stops", storeErr: errors.New("store failed")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls callRecorder
			cache := &cacheSpy{calls: &calls, deleteErr: tt.cacheErr}
			store := &storeSpy{calls: &calls, deleteErr: tt.storeErr}
			err := (&SessionService{cache: cache, repo: store}).Revoke(ctx, "token-hash")
			wantErr := tt.cacheErr
			if tt.storeErr != nil {
				wantErr = tt.storeErr
			}

			if !errors.Is(err, wantErr) {
				t.Fatalf("Revoke() error = %v, want %v", err, wantErr)
			}
			wantCalls := callRecorder{{method: "store.Delete", token: "token-hash"}}
			if tt.storeErr == nil {
				wantCalls = append(wantCalls, serviceCall{method: "cache.Delete", token: "token-hash"})
			}
			assertCalls(t, calls, wantCalls)
		})
	}
}
