package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"neomovies-api/pkg/middleware"
	"neomovies-api/pkg/models"
	"neomovies-api/pkg/services"
)

type ReactionsHandler struct {
	reactionsService *services.ReactionsService
}

func NewReactionsHandler(reactionsService *services.ReactionsService) *ReactionsHandler {
	return &ReactionsHandler{reactionsService: reactionsService}
}

func (h *ReactionsHandler) GetReactionCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mediaType := vars["mediaType"]
	mediaID := vars["mediaId"]

	if mediaType == "" || mediaID == "" {
		http.Error(w, "Media type and ID are required", http.StatusBadRequest)
		return
	}

	counts, err := h.reactionsService.GetReactionCounts(mediaType, mediaID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

func (h *ReactionsHandler) GetMyReaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	mediaType := vars["mediaType"]
	mediaID := vars["mediaId"]

	if mediaType == "" || mediaID == "" {
		http.Error(w, "Media type and ID are required", http.StatusBadRequest)
		return
	}

	reactionType, err := h.reactionsService.GetMyReaction(userID, mediaType, mediaID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if reactionType == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"type": reactionType})
	}
}

func (h *ReactionsHandler) SetReaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	mediaType := vars["mediaType"]
	mediaID := vars["mediaId"]

	if mediaType == "" || mediaID == "" {
		http.Error(w, "Media type and ID are required", http.StatusBadRequest)
		return
	}

	var request struct{ Type string `json:"type"` }
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if request.Type == "" {
		http.Error(w, "Reaction type is required", http.StatusBadRequest)
		return
	}

	if err := h.reactionsService.SetReaction(userID, mediaType, mediaID, request.Type); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Message: "Reaction set successfully"})
}

func (h *ReactionsHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	mediaType := vars["mediaType"]
	mediaID := vars["mediaId"]

	if mediaType == "" || mediaID == "" {
		http.Error(w, "Media type and ID are required", http.StatusBadRequest)
		return
	}

	if err := h.reactionsService.RemoveReaction(userID, mediaType, mediaID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Message: "Reaction removed successfully"})
}

func (h *ReactionsHandler) GetMyReactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	limit := getIntQuery(r, "limit", 50)

	reactions, err := h.reactionsService.GetUserReactions(userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: reactions})
}