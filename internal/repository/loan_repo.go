package repository

import (
	"context"
	"fmt"
	"home-library/internal/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LoanRepository struct {
	db *pgxpool.Pool
}

func NewLoanRepository(dbPool *pgxpool.Pool) *LoanRepository {
	return &LoanRepository{db: dbPool}
}

func (r *LoanRepository) Create(ctx context.Context, bookID int, borrower string, durationDays int) (int, error) {
	dueDate := time.Now().AddDate(0, 0, durationDays)

	query := `
			INSERT INTO loans (book_id, borrower_name, due_date)
			VALUES ($1, $2, $3)
			RETURNING id
	`
	var id int
	err := r.db.QueryRow(ctx, query, bookID, borrower, dueDate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to issue loan: %w", err)
	}

	return id, nil
}

func (r *LoanRepository) ReturnBook(ctx context.Context, loanID int) error {
	query := `
			UPDATE loans
			SET returned_at = NOW()
			WHERE id = $1 AND returned_at IS NULL		
	`

	cmdTag, err := r.db.Exec(ctx, query, loanID)
	if err != nil {
		return fmt.Errorf("failed to return book: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("аренда с ID %d не найдена или книга уже возвращена", loanID)
	}

	return nil
}

func (r *LoanRepository) GetActiveLoans(ctx context.Context, userID int) ([]models.Loan, error) {
	query := `
			SELECT l.id, l.book_id, b.title, l.borrower_name, l.loan_date, l.due_date
			FROM loans l
			JOIN books b ON l.book_id = b.id
			WHERE b.owner_id = $1 AND l.returned_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var loans []models.Loan

	for rows.Next() {
		var l models.Loan
		err := rows.Scan(&l.ID, &l.BookID, &l.BookTitle, &l.BorrowerName, &l.LoanDate, &l.DueDate)
		if err != nil {
			return nil, err
		}
		loans = append(loans, l)
	}
	return loans, nil
}
