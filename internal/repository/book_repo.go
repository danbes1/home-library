package repository

import (
	"context"
	"fmt"
	"home-library/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(dbPool *pgxpool.Pool) *BookRepository {
	return &BookRepository{db: dbPool}
}

func (r *BookRepository) GetAll(ctx context.Context, userID int) ([]models.Book, error) {
	query := `
		SELECT b.id, b.owner_id, b.title, b.authors, b.isbn, b.description, COALESCE(b.cover_url, ''), b.created_at 
		FROM books b
		JOIN users u ON b.owner_id = u.id
		WHERE u.id = $1 
		   OR (u.family_id IS NOT NULL AND u.family_id = (SELECT family_id FROM users WHERE id = $1))
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var books []models.Book

	for rows.Next() {

		var b models.Book
		err := rows.Scan(&b.ID, &b.OwnerID, &b.Title, &b.Authors, &b.ISBN, &b.Description, &b.CoverURL, &b.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		books = append(books, b)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return books, nil
}

func (r *BookRepository) Create(ctx context.Context, book models.Book) (int, error) {
	query := `
			INSERT INTO books (owner_id, title, authors, isbn, description, cover_url)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
	`
	var id int
	err := r.db.QueryRow(ctx, query, book.OwnerID, book.Title, book.Authors, book.ISBN, book.Description, book.CoverURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert book: %w", err)
	}
	return id, nil
}
