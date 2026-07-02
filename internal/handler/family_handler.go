package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"home-library/internal/repository"
)

type FamilyHandler struct {
	userRepo *repository.UserRepository
}

func NewFamilyHandler(ur *repository.UserRepository) *FamilyHandler {
	return &FamilyHandler{userRepo: ur}
}

func (h *FamilyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	familyName := strings.TrimSpace(r.FormValue("family_name"))
	if familyName == "" {
		http.Error(w, "Название семьи не может быть пустым", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	familyID, err := h.userRepo.CreateFamily(ctx, familyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.userRepo.UpdateFamilyID(ctx, userID, familyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *FamilyHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Не авторизован", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	familyIDStr := r.FormValue("family_id")
	familyID, err := strconv.Atoi(strings.TrimSpace(familyIDStr))
	if err != nil {
		http.Error(w, "Неверный формат ID семьи", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	exists, err := h.userRepo.CheckFamilyExists(ctx, familyID)
	if err != nil || !exists {
		http.Error(w, "Семья с таким кодом не найдена в системе", http.StatusNotFound)
		return
	}

	err = h.userRepo.UpdateFamilyID(ctx, userID, familyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
