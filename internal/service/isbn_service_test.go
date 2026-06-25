package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

func (m *MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

func TestFetchBookInfo_Success(t *testing.T) {
	fakeJSON := `{
		"items": [
			{
				"volumeInfo": {
					"title": "Тестовая книга",
					"authors": ["Петров Петр"],
					"description": "Классное описание книги"
				}
			}
		]
	}`

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(fakeJSON)),
	}

	svc := NewISBNService("fake_api_key")
	svc.client = &http.Client{
		Transport: &MockRoundTripper{Response: mockResponse},
	}

	ctx := context.Background()
	book, err := svc.FetchBookInfo(ctx, "9876543210")

	if err != nil {
		t.Fatalf("Ожидался успешный результат, а получили ошибку %v", err)
	}

	if book.Title != "Тестовая книга" {
		t.Errorf("Ожидалось название 'Тестовая книга', получили: %s", book.Title)
	}

	if len(book.Authors) != 1 || book.Authors[0] != "Петров Петр" {
		t.Errorf("Неверные авторы: %v", book.Authors)
	}
}

func TestFetchBookInfo_NotFound(t *testing.T) {

	fakeJSON := `{"items":[]}`

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(fakeJSON)),
	}

	svc := NewISBNService("fake_api_key")

	svc.client = &http.Client{
		Transport: &MockRoundTripper{Response: mockResponse},
	}

	svc.client = &http.Client{
		Transport: &MockRoundTripper{Response: mockResponse},
	}

	ctx := context.Background()
	_, err := svc.FetchBookInfo(ctx, "0000000000")

	if err == nil {
		t.Fatalf("Ожидалась ошибка 'книга не найдена', но метод завершился успешно")
	}
}
