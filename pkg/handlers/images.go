package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"neomovies-api/pkg/config"
)

type ImagesHandler struct{}

func NewImagesHandler() *ImagesHandler { return &ImagesHandler{} }

func (h *ImagesHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	size := vars["size"]
	imagePath := vars["path"]

	if size == "" || imagePath == "" {
		http.Error(w, "Size and path are required", http.StatusBadRequest)
		return
	}

	if imagePath == "placeholder.jpg" {
		h.servePlaceholder(w, r)
		return
	}

	validSizes := []string{"w92", "w154", "w185", "w342", "w500", "w780", "original"}
	if !h.isValidSize(size, validSizes) {
		size = "original"
	}

	imageURL := fmt.Sprintf("%s/%s/%s", config.TMDBImageBaseURL, size, imagePath)

	resp, err := http.Get(imageURL)
	if err != nil {
		h.servePlaceholder(w, r)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.servePlaceholder(w, r)
		return
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		h.servePlaceholder(w, r)
		return
	}
}

func (h *ImagesHandler) servePlaceholder(w http.ResponseWriter, r *http.Request) {
	placeholderPaths := []string{
		"./assets/placeholder.jpg",
		"./public/images/placeholder.jpg",
		"./static/placeholder.jpg",
	}

	var placeholderPath string
	for _, path := range placeholderPaths {
		if _, err := os.Stat(path); err == nil {
			placeholderPath = path
			break
		}
	}

	if placeholderPath == "" {
		h.serveSVGPlaceholder(w, r)
		return
	}

	file, err := os.Open(placeholderPath)
	if err != nil {
		h.serveSVGPlaceholder(w, r)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(placeholderPath))
	switch ext {
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		w.Header().Set("Content-Type", "image/jpeg")
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")

	_, err = io.Copy(w, file)
	if err != nil {
		h.serveSVGPlaceholder(w, r)
	}
}

func (h *ImagesHandler) serveSVGPlaceholder(w http.ResponseWriter, r *http.Request) {
	svgPlaceholder := `<svg width="300" height="450" xmlns="http://www.w3.org/2000/svg">
		<rect width="100%" height="100%" fill="#f0f0f0"/>
		<text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" font-family="Arial, sans-serif" font-size="16" fill="#666">
			Изображение не найдено
		</text>
	</svg>`

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(svgPlaceholder))
}

func (h *ImagesHandler) isValidSize(size string, validSizes []string) bool {
	for _, validSize := range validSizes {
		if size == validSize {
			return true
		}
	}
	return false
}