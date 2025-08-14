package models

type Movie struct {
	ID               int                    `json:"id"`
	Title            string                 `json:"title"`
	OriginalTitle    string                 `json:"original_title"`
	Overview         string                 `json:"overview"`
	PosterPath       string                 `json:"poster_path"`
	BackdropPath     string                 `json:"backdrop_path"`
	ReleaseDate      string                 `json:"release_date"`
	GenreIDs         []int                  `json:"genre_ids"`
	Genres           []Genre                `json:"genres"`
	VoteAverage      float64                `json:"vote_average"`
	VoteCount        int                    `json:"vote_count"`
	Popularity       float64                `json:"popularity"`
	Adult            bool                   `json:"adult"`
	Video            bool                   `json:"video"`
	OriginalLanguage string                 `json:"original_language"`
	Runtime          int                    `json:"runtime,omitempty"`
	Budget           int64                  `json:"budget,omitempty"`
	Revenue          int64                  `json:"revenue,omitempty"`
	Status           string                 `json:"status,omitempty"`
	Tagline          string                 `json:"tagline,omitempty"`
	Homepage         string                 `json:"homepage,omitempty"`
	IMDbID           string                 `json:"imdb_id,omitempty"`
	BelongsToCollection *Collection         `json:"belongs_to_collection,omitempty"`
	ProductionCompanies []ProductionCompany `json:"production_companies,omitempty"`
	ProductionCountries []ProductionCountry `json:"production_countries,omitempty"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages,omitempty"`
}

type TVShow struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	OriginalName     string                 `json:"original_name"`
	Overview         string                 `json:"overview"`
	PosterPath       string                 `json:"poster_path"`
	BackdropPath     string                 `json:"backdrop_path"`
	FirstAirDate     string                 `json:"first_air_date"`
	LastAirDate      string                 `json:"last_air_date"`
	GenreIDs         []int                  `json:"genre_ids"`
	Genres           []Genre                `json:"genres"`
	VoteAverage      float64                `json:"vote_average"`
	VoteCount        int                    `json:"vote_count"`
	Popularity       float64                `json:"popularity"`
	OriginalLanguage string                 `json:"original_language"`
	OriginCountry    []string               `json:"origin_country"`
	NumberOfEpisodes int                    `json:"number_of_episodes,omitempty"`
	NumberOfSeasons  int                    `json:"number_of_seasons,omitempty"`
	Status           string                 `json:"status,omitempty"`
	Type             string                 `json:"type,omitempty"`
	Homepage         string                 `json:"homepage,omitempty"`
	InProduction     bool                   `json:"in_production,omitempty"`
	Languages        []string               `json:"languages,omitempty"`
	Networks         []Network              `json:"networks,omitempty"`
	ProductionCompanies []ProductionCompany `json:"production_companies,omitempty"`
	ProductionCountries []ProductionCountry `json:"production_countries,omitempty"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages,omitempty"`
	CreatedBy           []Creator           `json:"created_by,omitempty"`
	EpisodeRunTime      []int               `json:"episode_run_time,omitempty"`
	Seasons             []Season            `json:"seasons,omitempty"`
}

// MultiSearchResult для мультипоиска
type MultiSearchResult struct {
	ID               int     `json:"id"`
	MediaType        string  `json:"media_type"` // "movie" или "tv"
	Title            string  `json:"title,omitempty"`       // для фильмов
	Name             string  `json:"name,omitempty"`        // для сериалов
	OriginalTitle    string  `json:"original_title,omitempty"`
	OriginalName     string  `json:"original_name,omitempty"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	ReleaseDate      string  `json:"release_date,omitempty"`     // для фильмов
	FirstAirDate     string  `json:"first_air_date,omitempty"`   // для сериалов
	GenreIDs         []int   `json:"genre_ids"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	Adult            bool    `json:"adult"`
	OriginalLanguage string  `json:"original_language"`
	OriginCountry    []string `json:"origin_country,omitempty"`
}

type MultiSearchResponse struct {
	Page         int                 `json:"page"`
	Results      []MultiSearchResult `json:"results"`
	TotalPages   int                 `json:"total_pages"`
	TotalResults int                 `json:"total_results"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GenresResponse struct {
	Genres []Genre `json:"genres"`
}

type ExternalIDs struct {
	ID          int    `json:"id"`
	IMDbID      string `json:"imdb_id"`
	TVDBID      int    `json:"tvdb_id,omitempty"`
	WikidataID  string `json:"wikidata_id"`
	FacebookID  string `json:"facebook_id"`
	InstagramID string `json:"instagram_id"`
	TwitterID   string `json:"twitter_id"`
}

type Collection struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

type ProductionCompany struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type ProductionCountry struct {
	ISO31661 string `json:"iso_3166_1"`
	Name     string `json:"name"`
}

type SpokenLanguage struct {
	EnglishName string `json:"english_name"`
	ISO6391     string `json:"iso_639_1"`
	Name        string `json:"name"`
}

type Network struct {
	ID            int    `json:"id"`
	LogoPath      string `json:"logo_path"`
	Name          string `json:"name"`
	OriginCountry string `json:"origin_country"`
}

type Creator struct {
	ID          int    `json:"id"`
	CreditID    string `json:"credit_id"`
	Name        string `json:"name"`
	Gender      int    `json:"gender"`
	ProfilePath string `json:"profile_path"`
}

type Season struct {
	AirDate      string `json:"air_date"`
	EpisodeCount int    `json:"episode_count"`
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}

type SeasonDetails struct {
	AirDate      string    `json:"air_date"`
	Episodes     []Episode `json:"episodes"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	ID           int       `json:"id"`
	PosterPath   string    `json:"poster_path"`
	SeasonNumber int       `json:"season_number"`
}

type Episode struct {
	AirDate       string  `json:"air_date"`
	EpisodeNumber int     `json:"episode_number"`
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	ProductionCode string `json:"production_code"`
	Runtime       int     `json:"runtime"`
	SeasonNumber  int     `json:"season_number"`
	ShowID        int     `json:"show_id"`
	StillPath     string  `json:"still_path"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
}

type TMDBResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

type TMDBTVResponse struct {
	Page         int      `json:"page"`
	Results      []TVShow `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

type SearchParams struct {
	Query        string `json:"query"`
	Page         int    `json:"page"`
	Language     string `json:"language"`
	Region       string `json:"region"`
	Year         int    `json:"year"`
	PrimaryReleaseYear int `json:"primary_release_year"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Модели для торрентов
type TorrentResult struct {
	Title       string    `json:"title"`
	Tracker     string    `json:"tracker"`
	Size        string    `json:"size"`
	Seeders     int       `json:"seeders"`
	Peers       int       `json:"peers"`
	Leechers    int       `json:"leechers"`
	Quality     string    `json:"quality"`
	Voice       []string  `json:"voice,omitempty"`
	Types       []string  `json:"types,omitempty"`
	Seasons     []int     `json:"seasons,omitempty"`
	Category    string    `json:"category"`
	MagnetLink  string    `json:"magnet"`
	TorrentLink string    `json:"torrent_link,omitempty"`
	Details     string    `json:"details,omitempty"`
	PublishDate string    `json:"publish_date"`
	AddedDate   string    `json:"added_date,omitempty"`
	Source      string    `json:"source"`
}

type TorrentSearchResponse struct {
	Query   string          `json:"query"`
	Results []TorrentResult `json:"results"`
	Total   int             `json:"total"`
}

// RedAPI специфичные структуры
type RedAPIResponse struct {
	Results []RedAPITorrent `json:"Results"`
}

type RedAPITorrent struct {
	Title       string            `json:"Title"`
	Tracker     string            `json:"Tracker"`
	Size        interface{}       `json:"Size"` // Может быть string или number
	Seeders     int               `json:"Seeders"`
	Peers       int               `json:"Peers"`
	MagnetUri   string            `json:"MagnetUri"`
	PublishDate string            `json:"PublishDate"`
	CategoryDesc string           `json:"CategoryDesc"`
	Details     string            `json:"Details"`
	Info        *RedAPITorrentInfo `json:"Info,omitempty"`
}

type RedAPITorrentInfo struct {
	Quality interface{} `json:"quality,omitempty"` // Может быть string или number
	Voices  []string    `json:"voices,omitempty"`
	Types   []string    `json:"types,omitempty"`
	Seasons []int       `json:"seasons,omitempty"`
}

// Alloha API структуры для получения информации о фильмах
type AllohaResponse struct {
	Data *AllohaData `json:"data"`
}

type AllohaData struct {
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
}

// Опции поиска торрентов
type TorrentSearchOptions struct {
	Season           *int
	Quality          []string
	MinQuality       string
	MaxQuality       string
	ExcludeQualities []string
	HDR              *bool
	HEVC             *bool
	SortBy           string
	SortOrder        string
	GroupByQuality   bool
	GroupBySeason    bool
	ContentType      string
}

// Модели для плееров
type PlayerResponse struct {
	Type   string `json:"type"`
	URL    string `json:"url"`
	Iframe string `json:"iframe,omitempty"`
}

// Модели для реакций
type Reaction struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	UserID   string `json:"userId" bson:"userId"`
	MediaID  string `json:"mediaId" bson:"mediaId"`
	Type     string `json:"type" bson:"type"`
	Created  string `json:"created" bson:"created"`
}

type ReactionCounts struct {
	Fire  int `json:"fire"`
	Nice  int `json:"nice"`
	Think int `json:"think"`
	Bore  int `json:"bore"`
	Shit  int `json:"shit"`
}