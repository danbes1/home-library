package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"home-library/internal/models"
	"home-library/internal/repository"
	"home-library/internal/service"
	"net/http"
	"time"
)

type BookHandler struct {
	bookRepo   *repository.BookRepository
	isbnSvc    *service.ISBNService
	barcodeSvc *service.BarcodeService
}

func NewBookHandler(br *repository.BookRepository, isbn *service.ISBNService, bar *service.BarcodeService) *BookHandler {
	return &BookHandler{bookRepo: br, isbnSvc: isbn, barcodeSvc: bar}
}

func (h *BookHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Пользователь не аутентифицирован"})
		return
	}

	books, err := h.bookRepo.GetAll(ctx, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Пользователь не аутентифицирован"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var newBook models.Book

	err := json.NewDecoder(r.Body).Decode(&newBook)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
		return
	}

	newBook.OwnerID = userID

	if newBook.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Поле title обязательное для заполнения"})
		return
	}

	id, err := h.bookRepo.Create(ctx, newBook)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	newBook.ID = id

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newBook)
}

func (h *BookHandler) Scan(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Файл слишком большой"})
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Поле 'image' обязательно для формы"})
		return
	}
	defer file.Close()

	isbnCode, err := h.barcodeSvc.ScanISBN(file)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	bookInfo, err := h.isbnSvc.FetchBookInfo(ctx, isbnCode)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"isbn":    isbnCode,
			"message": fmt.Sprintf("Код считан (%s), но данные о книге во внешних API не найдены", isbnCode),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"isbn":        isbnCode,
		"title":       bookInfo.Title,
		"authors":     bookInfo.Authors,
		"description": bookInfo.Description,
	})
}

func (h *BookHandler) ExternalSearchByISBN(w http.ResponseWriter, r *http.Request) {
	isbnCode := r.URL.Query().Get("code")
	if isbnCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Параметр 'code' обязателен"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 7*time.Second)
	defer cancel()

	bookInfo, err := h.isbnSvc.FetchBookInfo(ctx, isbnCode)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookInfo)
}
