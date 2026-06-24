package models

import "time"

type Loan struct {
	ID           int        `json:"id"`
	BookID       int        `json:"book_id"`
	BorrowerName string     `json:"borrower_name"`
	LoanDate     time.Time  `json:"loan_date"`
	DueDate      time.Time  `json:"due_date"`
	ReturnedAt   *time.Time `json:"returned_at,omitempty"`
}

type CreateLoanDTO struct {
	BookID       int    `json:""`
	BorrowerName string `json:"borrower_name"`
	DurationDays int    `json:"duration_days"`
}
