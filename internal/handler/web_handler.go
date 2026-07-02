package handler

import (
	"context"
	"home-library/internal/models"
	"home-library/internal/repository"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type WebHandler struct {
	bookRepo *repository.BookRepository
	userRepo *repository.UserRepository
	loanRepo *repository.LoanRepository
}

func NewWebHandler(br *repository.BookRepository, ur *repository.UserRepository, lr *repository.LoanRepository) *WebHandler {
	return &WebHandler{bookRepo: br, userRepo: ur, loanRepo: lr}
}

func (h *WebHandler) IndexPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		user = &models.User{ID: userID, Name: "Пользователь"}
	}

	books, err := h.bookRepo.GetAll(ctx, userID)
	if err != nil {
		books = []models.Book{}
	}

	loans, err := h.loanRepo.GetActiveLoans(ctx, userID)

	data := struct {
		User        *models.User
		Books       []models.Book
		ActiveLoans []models.Loan
	}{
		User:        user,
		Books:       books,
		ActiveLoans: loans,
	}

	basePath := filepath.Join("web", "templates", "base.html")
	pagePath := filepath.Join("web", "templates", "index.html")

	tmpl, err := template.ParseFiles(basePath, pagePath)
	if err != nil {
		http.Error(w, "Ошибка компиляции шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tmpl.ExecuteTemplate(w, "base", data)
}

func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("jwt_token"); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	pagePath := filepath.Join("web", "templates", "login.html")
	tmpl, err := template.ParseFiles(pagePath)
	if err != nil {
		http.Error(w, "Ошибка компиляции шаблона: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, nil)
}

func (h *WebHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("jwt_token"); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tmpl, _ := template.ParseFiles(filepath.Join("web", "templates", "register.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, nil)
}

func (h *WebHandler) AddBookPage(w http.ResponseWriter, r *http.Request) {
	basePath := filepath.Join("web", "templates", "base.html")
	pagePath := filepath.Join("web", "templates", "add-book.html")
	tmpl, _ := template.ParseFiles(basePath, pagePath)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "base", nil)
}
