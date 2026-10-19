# go-redis-session-lab

This lab explores Redis caching in Go through a small session lookup service. It compares
PostgreSQL-only lookups with Redis caching using HTTP benchmarks.

Motivation: [Let’s Handle 1 Million Requests per Second, It’s Scarier Than You Think!](https://youtu.be/W4EwfEU8CGA) by [Cododev](https://www.youtube.com/@Cododev).

## Reflection

To keep everything short, **this project was a semi-failure**. I could've chosen a more expensive
query to give caching more work to save on the PostgreSQL side. Instead, I benchmarked a
lightweight session lookup and got a modest improvement. I'm too tired to rebuild the workload
at this point just to get a more dramatic table.

The other main point of this lab was to get familiar with implementing Redis in Go. That part
worked, so I'm calling it good enough and moving on.

The query already has filters and a join, but they aren't particularly demanding. Now imagine
repeated requests involving more expensive filters and joins. A cache hit could skip more
database work, although how much that helps would need another benchmark.

## The Session Service

`GET /me` hashes a bearer token, looks up its session and user, and returns the user as JSON.
Both services use cache-aside: check the cache, fetch from Postgres on a miss, then populate
the cache. `SESSION_CACHE=redis` enables Redis caching; `postgres` uses no-op caches.
Session cache entries last up to five minutes, capped by the session's remaining lifetime.

## Running It

Check your available CPU cores and choose database pool sizes and benchmark concurrency
accordingly. Start small, monitor CPU and memory usage, and scale up only as far as your
machine can handle.

```bash
cp .env.example .env   # fill in DB_URL / SESSION_CACHE=redis|postgres / REDIS_*
docker compose up -d   # postgres + redis

# apply migrations
go tool goose -dir migrations postgres "$DB_URL" up

# seed some users + sessions to hit
psql "$DB_URL" -f scripts/seed.sql

go run ./cmd/server
curl -H "Authorization: Bearer benchmark-token-1" http://localhost:8080/me
```

## The Benchmark

The benchmark ran on two machines over Wi-Fi: one hosted the Go server, PostgreSQL, and Redis;
the other ran [k6](https://k6.io/) against `GET /me`. The database was seeded with 10M rows.
Each mode runs three passes of 100,000 requests at 25 concurrent users,
cycling through 1,000 bearer tokens. Each token is reused 100 times per pass.

`SESSION_CACHE=postgres` uses no cache; `redis` caches sessions and users. Redis is not flushed
between passes, and the cache hit rate was not measured.

## Results

| Mode            | Requests/sec | Avg latency |      p95 range |        p99 range | Failures |
| --------------- | -----------: | ----------: | -------------: | ---------------: | -------: |
| PostgreSQL only |        1,045 |    24.97 ms | 80.62–92.53 ms | 134.48–174.83 ms |        0 |
| Redis           |        1,145 |    22.84 ms | 69.19–86.69 ms | 128.76–142.17 ms |        0 |

Percentile ranges span the three passes; all 300,000 requests per mode returned HTTP 200.
Redis showed 9.6% higher throughput and 8.5% lower average latency. These timings include
network latency and measure performance at 25 concurrent users, not maximum capacity.
