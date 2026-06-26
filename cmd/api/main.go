package main

import (
	"context"
	"fmt"
	"home-library/internal/config"
	"home-library/internal/handler"
	"home-library/internal/repository"
	"home-library/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Printf("Не удалось подключиться к БД: %v\n", err)
		return
	}
	defer dbPool.Close()

	bookRepo := repository.NewBookRepository(dbPool)
	userRepo := repository.NewUserRepository(dbPool)
	loanRepo := repository.NewLoanRepository(dbPool)

	isbnSvc := service.NewISBNService(cfg.GoogleBooksAPIKey)
	barcodeSvc := service.NewBarcodeService()
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)

	if cfg.TelegramBotToken != "" {
		go func() {
			log.Printf("Фоновая инициализация Telegram-бота...\n")

			tgbot, err := service.NewTgBotService(cfg.TelegramBotToken, bookRepo, loanRepo, userRepo, isbnSvc, barcodeSvc)
			if err != nil {
				log.Printf("Ошибка инициализации Telegram-бота: %v\n", err)
				return
			}
			tgbot.Start()
		}()
	}

	bookHandler := handler.NewBookHandler(bookRepo, isbnSvc, barcodeSvc)
	authHandler := handler.NewAuthHandler(authSvc, userRepo, cfg.TelegramBotName)
	loanHandler := handler.NewLoanHandler(loanRepo)
	familyHandler := handler.NewFamilyHandler(userRepo)

	webHandler := handler.NewWebHandler(bookRepo, userRepo)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, req)
		})
	})

	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	r.Get("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/manifest.json")
	})
	r.Get("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/sw.js")
	})

	r.Get("/login", webHandler.LoginPage)
	r.Post("/login", authHandler.Login)
	r.Get("/register", webHandler.RegisterPage)
	r.Post("/register", authHandler.Register)

	r.Group(func(subRouter chi.Router) {
		subRouter.Use(handler.AuthMiddleware(cfg.JWTSecret))

		subRouter.Get("/", webHandler.IndexPage)

		subRouter.Get("/books", bookHandler.GetAll)
		subRouter.Post("/books", bookHandler.Create)
		subRouter.Get("/books/add", webHandler.AddBookPage)
		subRouter.Post("/books/scan", bookHandler.Scan)

		subRouter.Post("/loans", loanHandler.Create)
		subRouter.Get("/loans/active", loanHandler.GetActiveLoans)
		subRouter.Get("/loans/return", loanHandler.ReturnBook)

		subRouter.Post("/family/create", familyHandler.Create)
		subRouter.Post("/family/join", familyHandler.Join)

		subRouter.Post("/auth/tg-token", authHandler.GenerateTelegramToken)
	})

	fmt.Printf("Сервер запущен на порту %s\n", cfg.Port)

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %v\n", err)
	}
}
