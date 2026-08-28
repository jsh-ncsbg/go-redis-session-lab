package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/redis-session-lab/internal/db"
	"github.com/redis-session-lab/internal/session"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	cfg := initConfig()

	pg, err := db.NewPostgres(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal(fmt.Errorf("db setup: %w", err))
	}
	fmt.Println("connected to db")
	defer pg.Close()

	var cacheDB session.SessionCache
	switch strings.ToLower(cfg.SessionCache) {
	case "postgres":
		cacheDB = &session.NoopSessionCache{}
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
		cacheDB = session.NewRedisSessionCache(rdb, cfg.RedisPrefix)
	default:
		log.Fatalf("unsupported SESSION_CACHE value: %q", cfg.SessionCache)
	}

	sessionRepo := session.NewRepository(pg)
	sessionService := session.NewService(cacheDB, sessionRepo)

	mux := http.NewServeMux()
	mux.Handle("GET /me", meHandler(sessionService))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
