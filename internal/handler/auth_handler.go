package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"home-library/internal/repository"
	"home-library/internal/service"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AuthHandler struct {
	authSvc  *service.AuthService
	userRepo *repository.UserRepository
	botName  string
}

func NewAuthHandler(auth *service.AuthService, ur *repository.UserRepository, botName string) *AuthHandler {
	return &AuthHandler{authSvc: auth, userRepo: ur, botName: botName}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	id, err := h.authSvc.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"id": id, "message": "Регистрация успешна"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req.Email = r.FormValue("email")
		req.Password = r.FormValue("password")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	token, err := h.authSvc.Login(ctx, req.Email, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if r.Header.Get("Content-Type") != "application/json" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) GenerateTelegramToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Пользователь не аутентифицирован"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	token, err := uuid.NewRandom()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка генерации токена"})
		return
	}

	tokenStr := "auth_" + token.String()

	// 3. Сохраняем токен в базу данных для этого пользователя
	err = h.userRepo.SetAuthToken(ctx, userID, tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось сохранить токен в БД"})
		return
	}
	tgLink := fmt.Sprintf("https://t.me/%s/%s", h.botName, tokenStr)

	// 5. Возвращаем токен и готовую ссылку для фронтенда
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenStr,
		"link":  tgLink,
	})
}
