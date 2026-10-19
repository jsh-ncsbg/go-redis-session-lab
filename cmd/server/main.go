package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/redis-session-lab/internal/db"
	"github.com/redis-session-lab/internal/session"
	"github.com/redis-session-lab/internal/user"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	cfg := initConfig()

	pg, err := db.NewPostgres(ctx, cfg.DBURL, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxIdleTime)
	if err != nil {
		log.Fatal(fmt.Errorf("db setup: %w", err))
	}
	fmt.Println("connected to db")
	defer pg.Close()

	var sessionCache session.SessionCache
	var userCache user.UserCache
	switch strings.ToLower(cfg.SessionCache) {
	case "postgres":
		sessionCache = &session.NoopSessionCache{}
		userCache = &user.NoopUserCache{}
	case "redis":
		rdb := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		defer rdb.Close()

		if err := rdb.Ping(ctx).Err(); err != nil {
			panic(err)
		}

		fmt.Println("redis is available")
		sessionCache = session.NewRedisSessionCache(rdb, 5*time.Minute, cfg.RedisPrefix)
		userCache = user.NewRedisUserCache(rdb, 5*time.Minute, cfg.RedisPrefix)
	default:
		log.Fatalf("unsupported SESSION_CACHE value: %q", cfg.SessionCache)
	}

	sessionRepo := session.NewRepository(pg)
	sessionService := session.NewService(sessionCache, sessionRepo)

	userRepo := user.NewRepository(pg)
	userService := user.NewService(userCache, userRepo)

	mux := http.NewServeMux()
	mux.Handle("GET /me", meHandler(sessionService, userService))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
