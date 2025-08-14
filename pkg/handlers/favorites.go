package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"neomovies-api/pkg/middleware"
	"neomovies-api/pkg/models"
	"neomovies-api/pkg/services"
)

type FavoritesHandler struct {
	favoritesService *services.FavoritesService
}

func NewFavoritesHandler(favoritesService *services.FavoritesService) *FavoritesHandler {
	return &FavoritesHandler{
		favoritesService: favoritesService,
	}
}

func (h *FavoritesHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	favorites, err := h.favoritesService.GetFavorites(userID)
	if err != nil {
		http.Error(w, "Failed to get favorites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    favorites,
		Message: "Favorites retrieved successfully",
	})
}

func (h *FavoritesHandler) AddToFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	mediaID := vars["id"]
	mediaType := r.URL.Query().Get("type")
	
	if mediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		return
	}
	
	if mediaType == "" {
		mediaType = "movie" // По умолчанию фильм для обратной совместимости
	}
	
	if mediaType != "movie" && mediaType != "tv" {
		http.Error(w, "Media type must be 'movie' or 'tv'", http.StatusBadRequest)
		return
	}

	err := h.favoritesService.AddToFavorites(userID, mediaID, mediaType)
	if err != nil {
		http.Error(w, "Failed to add to favorites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Message: "Added to favorites successfully",
	})
}

func (h *FavoritesHandler) RemoveFromFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	mediaID := vars["id"]
	mediaType := r.URL.Query().Get("type")
	
	if mediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		return
	}
	
	if mediaType == "" {
		mediaType = "movie" // По умолчанию фильм для обратной совместимости
	}
	
	if mediaType != "movie" && mediaType != "tv" {
		http.Error(w, "Media type must be 'movie' or 'tv'", http.StatusBadRequest)
		return
	}

	err := h.favoritesService.RemoveFromFavorites(userID, mediaID, mediaType)
	if err != nil {
		http.Error(w, "Failed to remove from favorites: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Message: "Removed from favorites successfully",
	})
}

func (h *FavoritesHandler) CheckIsFavorite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	mediaID := vars["id"]
	mediaType := r.URL.Query().Get("type")
	
	if mediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		return
	}
	
	if mediaType == "" {
		mediaType = "movie" // По умолчанию фильм для обратной совместимости
	}
	
	if mediaType != "movie" && mediaType != "tv" {
		http.Error(w, "Media type must be 'movie' or 'tv'", http.StatusBadRequest)
		return
	}

	isFavorite, err := h.favoritesService.IsFavorite(userID, mediaID, mediaType)
	if err != nil {
		http.Error(w, "Failed to check favorite status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    map[string]bool{"isFavorite": isFavorite},
	})
}