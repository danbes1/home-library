package service

import (
	"context"
	"encoding/json"
	"fmt"
	"home-library/internal/models"
	"net/http"
	"time"
)

type ISBNService struct {
	client *http.Client
	apiKey string
}

func NewISBNService(apiKey string) *ISBNService {
	return &ISBNService{
		client: &http.Client{Timeout: 5 * time.Second},
		apiKey: apiKey,
	}
}

func (s *ISBNService) FetchBookInfo(ctx context.Context, isbn string) (*models.ExternalBook, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultChan := make(chan *models.ExternalBook, 2)
	errChan := make(chan error, 2)

	go func() {
		book, err := s.fetchOpenLibrary(ctx, isbn)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- book
	}()

	go func() {
		book, err := s.fetchGoogleBooks(ctx, isbn)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- book
	}()

	for i := 0; i < 2; i++ {
		select {
		case book := <-resultChan:
			cancel()
			return book, nil
		case <-errChan:
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("книга с ISBN %s не найдена ни в одном источнике", isbn)
}

func (s *ISBNService) fetchOpenLibrary(ctx context.Context, isbn string) (*models.ExternalBook, error) {
	url := fmt.Sprintf("https://openlibrary.org/isbn/%s", isbn)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "HomeLibraryMicroservice/1.0 (assasinhak008@gmail.com)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openlibrary status: %d", resp.StatusCode)
	}

	var result map[string]struct {
		Title   string `json:"title"`
		Authors []struct {
			Name string `json:"name"`
		} `json:"authors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	bookData, exists := result["ISBN:"+isbn]
	if !exists {
		return nil, fmt.Errorf("not found")
	}

	var authors []string
	for _, a := range bookData.Authors {
		authors = append(authors, a.Name)
	}

	return &models.ExternalBook{
		Title:   bookData.Title,
		Authors: authors,
	}, nil
}

func (s *ISBNService) fetchGoogleBooks(ctx context.Context, isbn string) (*models.ExternalBook, error) {
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=isbn:%s&key=%s", isbn, s.apiKey)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "HomeLibraryMicroservice/1.0 (assasinhak008@gmail.com)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google books status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			VolumeInfo struct {
				Title       string   `json:"title"`
				Authors     []string `json:"authors"`
				Description string   `json:"description"`
			} `json:"volumeInfo"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("not found")
	}

	info := result.Items[0].VolumeInfo

	return &models.ExternalBook{
		Title:       info.Title,
		Authors:     info.Authors,
		Description: info.Description,
	}, nil
}
