package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"neomovies-api/pkg/config"
	"neomovies-api/pkg/database"
	appHandlers "neomovies-api/pkg/handlers"
	"neomovies-api/pkg/middleware"
	"neomovies-api/pkg/monitor"
	"neomovies-api/pkg/services"
)

func main() {
	if err := godotenv.Load(); err != nil { _ = err }

	cfg := config.New()

	db, err := database.Connect(cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Disconnect()

	tmdbService := services.NewTMDBService(cfg.TMDBAccessToken)
	emailService := services.NewEmailService(cfg)
	authService := services.NewAuthService(db, cfg.JWTSecret, emailService, cfg.BaseURL, cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL, cfg.FrontendURL)

	movieService := services.NewMovieService(db, tmdbService)
	tvService := services.NewTVService(db, tmdbService)
	torrentService := services.NewTorrentServiceWithConfig(cfg.RedAPIBaseURL, cfg.RedAPIKey)
	reactionsService := services.NewReactionsService(db)

	authHandler := appHandlers.NewAuthHandler(authService)
	movieHandler := appHandlers.NewMovieHandler(movieService)
	tvHandler := appHandlers.NewTVHandler(tvService)
	docsHandler := appHandlers.NewDocsHandler()
	searchHandler := appHandlers.NewSearchHandler(tmdbService)
	categoriesHandler := appHandlers.NewCategoriesHandler(tmdbService)
	playersHandler := appHandlers.NewPlayersHandler(cfg)
	torrentsHandler := appHandlers.NewTorrentsHandler(torrentService, tmdbService)
	reactionsHandler := appHandlers.NewReactionsHandler(reactionsService)
	imagesHandler := appHandlers.NewImagesHandler()

	r := mux.NewRouter()

	r.HandleFunc("/", docsHandler.ServeDocs).Methods("GET")
	r.HandleFunc("/openapi.json", docsHandler.GetOpenAPISpec).Methods("GET")

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", appHandlers.HealthCheck).Methods("GET")
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/verify", authHandler.VerifyEmail).Methods("POST")
	api.HandleFunc("/auth/resend-code", authHandler.ResendVerificationCode).Methods("POST")
	api.HandleFunc("/auth/google/login", authHandler.GoogleLogin).Methods("GET")
	api.HandleFunc("/auth/google/callback", authHandler.GoogleCallback).Methods("GET")

	api.HandleFunc("/search/multi", searchHandler.MultiSearch).Methods("GET")

	api.HandleFunc("/categories", categoriesHandler.GetCategories).Methods("GET")
	api.HandleFunc("/categories/{id}/movies", categoriesHandler.GetMoviesByCategory).Methods("GET")

	api.HandleFunc("/players/alloha/{imdb_id}", playersHandler.GetAllohaPlayer).Methods("GET")
	api.HandleFunc("/players/lumex/{imdb_id}", playersHandler.GetLumexPlayer).Methods("GET")
    api.HandleFunc("/players/vibix/{imdb_id}", playersHandler.GetVibixPlayer).Methods("GET")

	api.HandleFunc("/torrents/search/{imdbId}", torrentsHandler.SearchTorrents).Methods("GET")
	api.HandleFunc("/torrents/movies", torrentsHandler.SearchMovies).Methods("GET")
	api.HandleFunc("/torrents/series", torrentsHandler.SearchSeries).Methods("GET")
	api.HandleFunc("/torrents/anime", torrentsHandler.SearchAnime).Methods("GET")
	api.HandleFunc("/torrents/seasons", torrentsHandler.GetAvailableSeasons).Methods("GET")
	api.HandleFunc("/torrents/search", torrentsHandler.SearchByQuery).Methods("GET")

	api.HandleFunc("/reactions/{mediaType}/{mediaId}/counts", reactionsHandler.GetReactionCounts).Methods("GET")

	api.HandleFunc("/images/{size}/{path:.*}", imagesHandler.GetImage).Methods("GET")

	api.HandleFunc("/movies/search", movieHandler.Search).Methods("GET")
	api.HandleFunc("/movies/popular", movieHandler.Popular).Methods("GET")
	api.HandleFunc("/movies/top-rated", movieHandler.TopRated).Methods("GET")
	api.HandleFunc("/movies/upcoming", movieHandler.Upcoming).Methods("GET")
	api.HandleFunc("/movies/now-playing", movieHandler.NowPlaying).Methods("GET")
	api.HandleFunc("/movies/{id}", movieHandler.GetByID).Methods("GET")
	api.HandleFunc("/movies/{id}/recommendations", movieHandler.GetRecommendations).Methods("GET")
	api.HandleFunc("/movies/{id}/similar", movieHandler.GetSimilar).Methods("GET")
	api.HandleFunc("/movies/{id}/external-ids", movieHandler.GetExternalIDs).Methods("GET")

	api.HandleFunc("/tv/search", tvHandler.Search).Methods("GET")
	api.HandleFunc("/tv/popular", tvHandler.Popular).Methods("GET")
	api.HandleFunc("/tv/top-rated", tvHandler.TopRated).Methods("GET")
	api.HandleFunc("/tv/on-the-air", tvHandler.OnTheAir).Methods("GET")
	api.HandleFunc("/tv/airing-today", tvHandler.AiringToday).Methods("GET")
	api.HandleFunc("/tv/{id}", tvHandler.GetByID).Methods("GET")
	api.HandleFunc("/tv/{id}/recommendations", tvHandler.GetRecommendations).Methods("GET")
	api.HandleFunc("/tv/{id}/similar", tvHandler.GetSimilar).Methods("GET")
	api.HandleFunc("/tv/{id}/external-ids", tvHandler.GetExternalIDs).Methods("GET")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))

	protected.HandleFunc("/favorites", movieHandler.GetFavorites).Methods("GET")
	protected.HandleFunc("/favorites/{id}", movieHandler.AddToFavorites).Methods("POST")
	protected.HandleFunc("/favorites/{id}", movieHandler.RemoveFromFavorites).Methods("DELETE")

	protected.HandleFunc("/auth/profile", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/auth/profile", authHandler.UpdateProfile).Methods("PUT")
	protected.HandleFunc("/auth/profile", authHandler.DeleteAccount).Methods("DELETE")

	protected.HandleFunc("/reactions/{mediaType}/{mediaId}/my-reaction", reactionsHandler.GetMyReaction).Methods("GET")
	protected.HandleFunc("/reactions/{mediaType}/{mediaId}", reactionsHandler.SetReaction).Methods("POST")
	protected.HandleFunc("/reactions/{mediaType}/{mediaId}", reactionsHandler.RemoveReaction).Methods("DELETE")
	protected.HandleFunc("/reactions/my", reactionsHandler.GetMyReactions).Methods("GET")

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Accept", "Origin", "X-Requested-With", "X-CSRF-Token"}),
		handlers.AllowCredentials(),
		handlers.ExposedHeaders([]string{"Authorization", "Content-Type"}),
	)

	var finalHandler http.Handler
	if cfg.NodeEnv == "development" {
		r.Use(monitor.RequestMonitor())
		finalHandler = corsHandler(r)

		fmt.Println("\n🚀 NeoMovies API Server")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📡 Server: http://localhost:%s\n", cfg.Port)
		fmt.Printf("📚 Docs:   http://localhost:%s/\n", cfg.Port)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("%-6s %-3s │ %-60s │ %8s\n", "METHOD", "CODE", "ENDPOINT", "TIME")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	} else {
		finalHandler = corsHandler(r)
		fmt.Printf("✅ Server starting on port %s\n", cfg.Port)
	}

	port := cfg.Port
	if port == "" { port = "3000" }

	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
		os.Exit(1)
	}
}