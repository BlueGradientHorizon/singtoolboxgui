package common

import (
	"fmt"
	"image"
	"image/color"

	"github.com/yeqown/go-qrcode/v2"
)

// MatrixWriter is a custom writer that captures the QR code matrix
type MatrixWriter struct {
	Matrix qrcode.Matrix
}

func (w *MatrixWriter) Write(mat qrcode.Matrix) error {
	w.Matrix = mat
	return nil
}

func (w *MatrixWriter) Close() error {
	return nil
}

func GenerateQRCode(text string) (image.Image, error) {
	qr, err := qrcode.New(text)
	if err != nil {
		return nil, err
	}

	mw := &MatrixWriter{}
	if err := qr.Save(mw); err != nil {
		return nil, err
	}

	bitmap := mw.Matrix.Bitmap()
	if len(bitmap) == 0 {
		return nil, fmt.Errorf("empty QR code bitmap")
	}

	bitmapSize := len(bitmap)
	if bitmapSize == 0 {
		return nil, fmt.Errorf("empty QR code bitmap")
	}

	// Add padding (quiet zone) around QR code - 4 modules on each side
	padding := 4
	totalModules := bitmapSize + padding*2

	img := image.NewRGBA(image.Rect(0, 0, totalModules, totalModules))

	// Fill the entire image with white (background + quiet zone)
	for y := range totalModules {
		for x := range totalModules {
			img.Set(x, y, color.White)
		}
	}

	for y := range bitmapSize {
		for x := range bitmapSize {
			// Only draw black modules (white is already the background)
			if y < len(bitmap) && x < len(bitmap[y]) && bitmap[y][x] {
				startX := (x + padding)
				startY := (y + padding)
				img.Set(startX, startY, color.Black)
			}
		}
	}

	return img, nil
}
