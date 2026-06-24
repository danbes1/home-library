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

func (r *BookRepository) GetAll(ctx context.Context) ([]models.Book, error) {
	query := `SELECT id, owner_id, title, authors, isbn, description, created_at FROM books`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var books []models.Book

	for rows.Next() {

		var b models.Book
		err := rows.Scan(&b.ID, &b.OwnerID, &b.Title, &b.Authors, &b.ISBN, &b.Description, &b.CreatedAt)
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
			INSERT INTO books (owner_id, title, authors, isbn, description)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
	`
	var id int
	err := r.db.QueryRow(ctx, query, book.OwnerID, book.Title, book.Authors, book.ISBN, book.Description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert book: %w", err)
	}
	return id, nil
}
