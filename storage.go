package main

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var ErrNoFile = errors.New("no file")

// saveUploadedFile saves a file from a multipart form field to uploads directory.
// It returns the relative path (filename only) or empty string if no file was provided.
func saveUploadedFile(r *http.Request, field string) (string, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", ErrNoFile
		}
		var maxErr *multipart.Part
		if errors.As(err, &maxErr) {
			return "", fmt.Errorf("file too large")
		}
		return "", err
	}
	defer file.Close()

	if header.Filename == "" {
		return "", ErrNoFile
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".dat"
	}
	filename := uuid.NewString() + ext
	dstPath := filepath.Join(UploadsDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filename, nil
}


