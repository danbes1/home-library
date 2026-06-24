package service

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
)

type BarcodeService struct{}

func NewBarcodeService() *BarcodeService {
	return &BarcodeService{}
}

func (s *BarcodeService) ScanISBN(fileReader io.Reader) (string, error) {
	img, _, err := image.Decode(fileReader)
	if err != nil {
		return "", fmt.Errorf("Не удалось декодировать изображение: %w", err)
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("ошибка обработки растра: %w", err)
	}

	reader := oned.NewEAN13Reader()

	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("штрихкод не распознан, попробуйте сделать фото чётче: %w", err)
	}

	isbn := result.GetText()
	if len(isbn) == 0 {
		return "", fmt.Errorf("код распознан как пустой")
	}

	return isbn, nil
}
