package repository

import (
	"context"
	"fmt"
	"home-library/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepoInterface interface {
	Create(ctx context.Context, name, email, passwordHash string) (int, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: dbPool,
	}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (int, error) {
	query := `
			INSERT INTO users (name, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id
	`

	var id int
	err := r.db.QueryRow(ctx, query, name, email, passwordHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID int) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, family_id, created_at FROM users WHERE id = $1`

	var u models.User

	err := r.db.QueryRow(ctx, query, userID).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.FamilyID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`

	var u models.User

	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByTelegramChatID(ctx context.Context, chatID int64) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, family_id, created_at FROM users WHERE telegram_chat_id = $1`
	var u models.User

	err := r.db.QueryRow(ctx, query, chatID).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.FamilyID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) SetAuthToken(ctx context.Context, userID int, token string) error {
	query := `UPDATE users SET telegram_auth_token = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, token, userID)
	return err
}

func (r *UserRepository) LinkTelegramByToken(ctx context.Context, token string, chatID int64) (*models.User, error) {
	query := `
		UPDATE users 
		SET telegram_chat_id = $1, telegram_auth_token = NULL 
		WHERE telegram_auth_token = $2
		RETURNING id, name, email, family_id, telegram_chat_id, created_at
	`
	var u models.User
	err := r.db.QueryRow(ctx, query, chatID, token).Scan(&u.ID, &u.Name, &u.Email, &u.FamilyID, &u.TelegramChatID, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("неверный токен или аккаунт уже привязан: %w", err)
	}
	return &u, nil
}
