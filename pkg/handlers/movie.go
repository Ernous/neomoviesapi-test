package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"neomovies-api/pkg/models"
	"neomovies-api/pkg/services"
)

type MovieHandler struct {
	movieService *services.MovieService
}

func NewMovieHandler(movieService *services.MovieService) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
	}
}

func (h *MovieHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")
	region := r.URL.Query().Get("region")
	year := getIntQuery(r, "year", 0)

	movies, err := h.movieService.Search(query, page, language, region, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}

func (h *MovieHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	language := r.URL.Query().Get("language")

	movie, err := h.movieService.GetByID(id, language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movie,
	})
}

func (h *MovieHandler) Popular(w http.ResponseWriter, r *http.Request) {
	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")
	region := r.URL.Query().Get("region")

	movies, err := h.movieService.GetPopular(page, language, region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}

func (h *MovieHandler) TopRated(w http.ResponseWriter, r *http.Request) {
	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")
	region := r.URL.Query().Get("region")

	movies, err := h.movieService.GetTopRated(page, language, region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}

func (h *MovieHandler) Upcoming(w http.ResponseWriter, r *http.Request) {
	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")
	region := r.URL.Query().Get("region")

	movies, err := h.movieService.GetUpcoming(page, language, region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}

func (h *MovieHandler) NowPlaying(w http.ResponseWriter, r *http.Request) {
	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")
	region := r.URL.Query().Get("region")

	movies, err := h.movieService.GetNowPlaying(page, language, region)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}

func (h *MovieHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")

	movies, err := h.movieService.GetRecommendations(id, page, language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}

func (h *MovieHandler) GetSimilar(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	page := getIntQuery(r, "page", 1)
	language := r.URL.Query().Get("language")

	movies, err := h.movieService.GetSimilar(id, page, language)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    movies,
	})
}



func (h *MovieHandler) GetExternalIDs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	externalIDs, err := h.movieService.GetExternalIDs(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    externalIDs,
	})
}

func getIntQuery(r *http.Request, key string, defaultValue int) int {
	str := r.URL.Query().Get(key)
	if str == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}
	
	return value
}