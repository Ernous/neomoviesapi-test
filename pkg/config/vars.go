package config

const (
	// Environment variable keys
	EnvTMDBAccessToken   = "TMDB_ACCESS_TOKEN"
	EnvJWTSecret         = "JWT_SECRET"
	EnvPort              = "PORT"
	EnvBaseURL           = "BASE_URL"
	EnvNodeEnv           = "NODE_ENV"
	EnvGmailUser         = "GMAIL_USER"
	EnvGmailPassword     = "GMAIL_APP_PASSWORD"
	EnvLumexURL          = "LUMEX_URL"
	EnvAllohaToken       = "ALLOHA_TOKEN"
	EnvRedAPIBaseURL     = "REDAPI_BASE_URL"
	EnvRedAPIKey         = "REDAPI_KEY"
	EnvMongoDBName       = "MONGO_DB_NAME"
	EnvGoogleClientID    = "GOOGLE_CLIENT_ID"
	EnvGoogleClientSecret= "GOOGLE_CLIENT_SECRET"
	EnvGoogleRedirectURL = "GOOGLE_REDIRECT_URL"
	EnvFrontendURL       = "FRONTEND_URL"

	// Default values
	DefaultJWTSecret   = "your-secret-key"
	DefaultPort        = "3000"
	DefaultBaseURL     = "http://localhost:3000"
	DefaultNodeEnv     = "development"
	DefaultRedAPIBase  = "http://redapi.cfhttp.top"
	DefaultMongoDBName = "database"

	// Static constants
	TMDBImageBaseURL = "https://image.tmdb.org/t/p"
	CubAPIBaseURL    = "https://cub.rip/api"
)