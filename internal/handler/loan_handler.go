package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"home-library/internal/models"
	"home-library/internal/repository"
	"net/http"
	"time"
)

type LoanHandler struct {
	loanRepo *repository.LoanRepository
}

func NewLoanHandler(loan *repository.LoanRepository) *LoanHandler {
	return &LoanHandler{loanRepo: loan}
}

func (h *LoanHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var dto models.CreateLoanDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}
	if dto.DurationDays <= 0 {
		dto.DurationDays = 14
	}

	id, err := h.loanRepo.Create(ctx, dto.BookID, dto.BorrowerName, dto.DurationDays)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"loan_id": id, "message": "Книга успешно передана другу"})
}

func (h *LoanHandler) GetActiveLoans(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUserIDFromContext(r.Context())
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	loans, err := h.loanRepo.GetActiveLoans(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loans)
}

func (h *LoanHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	loanIDStr := r.URL.Query().Get("id")
	if loanIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Параметр id обязателен"})
		return
	}

	var loanID int

	_, err := fmt.Sscanf(loanIDStr, "%d", &loanID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат ID"})
		return
	}

	err = h.loanRepo.ReturnBook(ctx, loanID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Книга успешно возвращена"})
}
