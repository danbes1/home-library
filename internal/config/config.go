package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	GoogleBooksAPIKey string
	JWTSecret         string
	TelegramBotToken  string
	TelegramBotName   string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Предупреждение: .env файл не найден")
	}

	return &Config{
		Port:              getEnv("PORT", ":80"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/home_library?sslmode=disable"),
		GoogleBooksAPIKey: getEnv("GOOGLE_BOOKS_API_KEY", "<PASTE-YOUR-API-KEY>"),
		JWTSecret:         getEnv("JWT_SECRET", "<PASTE-YOUR-JWT-SECRET>"),
		TelegramBotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramBotName:   getEnv("TELEGRAM_BOT_NAME", ""),
	}
}

func getEnv(key, fallback string) string {
	value := fallback
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return os.ExpandEnv(value)
}
