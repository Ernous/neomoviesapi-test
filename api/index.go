package handler

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"

    "neomovies-api/pkg/config"
    "neomovies-api/pkg/database"
    handlersPkg "neomovies-api/pkg/handlers"
    "neomovies-api/pkg/middleware"
    "neomovies-api/pkg/services"
)

var (
    globalDB  *mongo.Database
    globalCfg *config.Config
    initOnce  sync.Once
    initError error
)

func initializeApp() {
    if err := godotenv.Load(); err != nil { _ = err }

    globalCfg = config.New()

    var err error
    globalDB, err = database.Connect(globalCfg.MongoURI, globalCfg.MongoDBName)
    if err != nil {
        log.Printf("Failed to connect to database: %v", err)
        initError = err
        return
    }

    log.Println("Successfully connected to database")
}

func Handler(w http.ResponseWriter, r *http.Request) {
    initOnce.Do(initializeApp)

    if initError != nil {
        log.Printf("Initialization error: %v", initError)
        http.Error(w, "Application initialization failed: "+initError.Error(), http.StatusInternalServerError)
        return
    }

    tmdbService := services.NewTMDBService(globalCfg.TMDBAccessToken)
    emailService := services.NewEmailService(globalCfg)
    authService := services.NewAuthService(globalDB, globalCfg.JWTSecret, emailService, globalCfg.BaseURL, globalCfg.GoogleClientID, globalCfg.GoogleClientSecret, globalCfg.GoogleRedirectURL, globalCfg.FrontendURL)

    movieService := services.NewMovieService(globalDB, tmdbService)
    tvService := services.NewTVService(globalDB, tmdbService)
    torrentService := services.NewTorrentServiceWithConfig(globalCfg.RedAPIBaseURL, globalCfg.RedAPIKey)
    reactionsService := services.NewReactionsService(globalDB)

    authHandler := handlersPkg.NewAuthHandler(authService)
    movieHandler := handlersPkg.NewMovieHandler(movieService)
    tvHandler := handlersPkg.NewTVHandler(tvService)
    docsHandler := handlersPkg.NewDocsHandler()
    searchHandler := handlersPkg.NewSearchHandler(tmdbService)
    categoriesHandler := handlersPkg.NewCategoriesHandler(tmdbService)
    playersHandler := handlersPkg.NewPlayersHandler(globalCfg)
    torrentsHandler := handlersPkg.NewTorrentsHandler(torrentService, tmdbService)
    reactionsHandler := handlersPkg.NewReactionsHandler(reactionsService)
    imagesHandler := handlersPkg.NewImagesHandler()

    router := mux.NewRouter()

    router.HandleFunc("/", docsHandler.ServeDocs).Methods("GET")
    router.HandleFunc("/openapi.json", docsHandler.GetOpenAPISpec).Methods("GET")

    api := router.PathPrefix("/api/v1").Subrouter()

    api.HandleFunc("/health", handlersPkg.HealthCheck).Methods("GET")
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
    protected.Use(middleware.JWTAuth(globalCfg.JWTSecret))

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

    corsHandler(router).ServeHTTP(w, r)
}