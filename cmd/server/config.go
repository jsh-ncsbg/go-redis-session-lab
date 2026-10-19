package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL             string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxIdleTime time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisPrefix   string

	SessionCache string
}

func initConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		DBURL:             GetEnvString("DB_URL", "postgres://postgres:postgres@localhost:5432/rsl?sslmode=disable"),
		DBMaxOpenConns:    GetEnvInt("DB_MAX_OPEN_CONNS", 30),
		DBMaxIdleConns:    GetEnvInt("DB_MAX_IDLE_CONNS", 30),
		DBConnMaxIdleTime: GetEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),

		RedisAddr:     GetEnvString("REDIS_ADDR", "localhost:6379"),
		RedisPassword: GetEnvString("REDIS_PASSWORD", ""),
		RedisDB:       GetEnvInt("REDIS_DB", 0),
		RedisPrefix:   GetEnvString("REDIS_PREFIX", "rsl"),

		SessionCache: GetEnvString("SESSION_CACHE", "postgres"),
	}
}

func GetEnvString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	val = strings.TrimSpace(val)
	if !ok || val == "" {
		return fallback
	}
	return val
}

func GetEnvInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	val = strings.TrimSpace(val)
	if !ok || val == "" {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}

func GetEnvDuration(key string, fallback time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	val = strings.TrimSpace(val)
	if !ok || val == "" {
		return fallback
	}
	valAsDuration, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return valAsDuration
}
