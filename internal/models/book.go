package models

import "time"

type Book struct {
	ID          int       `json:"id"`
	OwnerID     int       `json:"owner_id"`
	Title       string    `json:"title"`
	Authors     []string  `json:"authors"`
	ISBN        string    `json:"isbn,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateBookDTO struct {
	Title   string   `json:"title"`
	Authors []string `json:"authors"`
	ISBN    string   `json:"isbn"`
}
