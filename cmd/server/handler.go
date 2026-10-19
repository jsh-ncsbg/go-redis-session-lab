package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/redis-session-lab/internal/session"
	"github.com/redis-session-lab/internal/user"
)

func meHandler(sessionService *session.SessionService, userService *user.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		sum := sha256.Sum256([]byte(parts[1]))
		tokenHash := hex.EncodeToString(sum[:])

		sess, err := sessionService.Get(r.Context(), tokenHash)
		if err != nil {
			if errors.Is(err, session.ErrSessionNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		usr, err := userService.GetUserByID(r.Context(), sess.User.ID)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(usr)
	}
}
