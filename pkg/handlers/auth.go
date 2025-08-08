package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"neomovies-api/pkg/middleware"
	"neomovies-api/pkg/models"
	"neomovies-api/pkg/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Register(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: response, Message: "User registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "Account not activated. Please verify your email." {
			statusCode = http.StatusForbidden
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: response, Message: "Login successful"})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, HttpOnly: true, Path: "/", Expires: time.Now().Add(10 * time.Minute)})
	url, err := h.authService.GetGoogleLoginURL(state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	state := q.Get("state")
	code := q.Get("code") 
	preferJSON := q.Get("response") == "json" || strings.Contains(r.Header.Get("Accept"), "application/json")
	cookie, _ := r.Cookie("oauth_state")
	if cookie == nil || cookie.Value != state || code == "" {
		if preferJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.APIResponse{Success: false, Message: "invalid oauth state"})
			return
		}
		redirectURL, ok := h.authService.BuildFrontendRedirect("", "invalid_state")
		if ok {
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	resp, err := h.authService.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		if preferJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.APIResponse{Success: false, Message: err.Error()})
			return
		}
		redirectURL, ok := h.authService.BuildFrontendRedirect("", "auth_failed")
		if ok {
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if preferJSON {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: resp, Message: "Login successful"})
		return
	}

	redirectURL, ok := h.authService.BuildFrontendRedirect(resp.Token, "")
	if ok {
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: resp, Message: "Login successful"})
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: user})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	delete(updates, "password")
	delete(updates, "email")
	delete(updates, "_id")
	delete(updates, "created_at")

	user, err := h.authService.UpdateUser(userID, bson.M(updates))
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: user, Message: "Profile updated successfully"})
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	if err := h.authService.DeleteAccount(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Message: "Account deleted successfully"})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.VerifyEmail(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) ResendVerificationCode(w http.ResponseWriter, r *http.Request) {
	var req models.ResendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.ResendVerificationCode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// helpers
func generateState() string { return uuidNew() }