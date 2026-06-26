package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			if tokenString == "" {
				cookie, err := r.Cookie("jwt_token")
				if err == nil {
					tokenString = cookie.Value
				}
			}
			if tokenString == "" {
				if strings.Contains(r.Header.Get("Accept"), "text/html") {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Отсутствует токен авторизации"})
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				if strings.Contains(r.Header.Get("Accept"), "text/html") {
					http.SetCookie(w, &http.Cookie{
						Name:     "jwt_token",
						Value:    "",
						Path:     "/",
						MaxAge:   -1,
						HttpOnly: true,
					})
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Недействительный или просроченный токен"})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userID := int(userIDFloat)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
