package service

import (
	"context"
	"errors"
	"home-library/internal/models"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	users  map[string]*models.User
	lastID int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, name, email, passwordHash string) (int, error) {
	m.lastID++
	m.users[email] = &models.User{
		ID:           m.lastID,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	return m.lastID, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user, exist := m.users[email]
	if !exist {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := NewMockUserRepository()
	svc := NewAuthService(mockRepo, "test_secret")

	ctx := context.Background()
	userID, err := svc.Register(ctx, "Даниил", "dan@example.com", "secret_pass")

	if err != nil {
		t.Fatalf("Ожидалась успешная регистрация, но получена ошибка: %v", err)
	}

	if userID != 1 {
		t.Errorf("Ожидался ID = 1, но получен: %d", userID)
	}

	savedUser := mockRepo.users["dan@example.com"]
	if savedUser.Name != "Даниил" {
		t.Errorf("Ожидалось имя 'Даниил', получено: '%s'", savedUser.Name)
	}

	if savedUser.PasswordHash == "secret_pass" {
		t.Error("Пароль сохранился в открытом виде! Хэширование не сработало.")
	}

	err = bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte("secret_pass"))
	if err != nil {
		t.Errorf("Пароль не соответствует созданному хэшу: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo := NewMockUserRepository()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), 10)
	_, _ = mockRepo.Create(context.Background(), "Не Даниил", "notdan@example.com", string(hash))

	jwtSecret := "super_secret_key"
	svc := NewAuthService(mockRepo, jwtSecret)

	// Тестируем вход
	tokenString, err := svc.Login(context.Background(), "notdan@example.com", "correct_password")
	if err != nil {
		t.Fatalf("Ожидался успешный вход, но получена ошибка: %v", err)
	}

	if tokenString == "" {
		t.Fatal("Метод вернул пустой токен")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		t.Fatalf("Сгенерированный токен невалиден: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Не удалось прочитать Claims из токена")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok || int(userIDFloat) != 1 {
		t.Errorf("В токене сохранен неверный user_id: %v", claims["user_id"])
	}
}
