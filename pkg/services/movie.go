package services

import (
	"go.mongodb.org/mongo-driver/mongo"

	"neomovies-api/pkg/models"
)

type MovieService struct {
	tmdb *TMDBService
}

func NewMovieService(db *mongo.Database, tmdb *TMDBService) *MovieService {
	return &MovieService{
		tmdb: tmdb,
	}
}

func (s *MovieService) Search(query string, page int, language, region string, year int) (*models.TMDBResponse, error) {
	return s.tmdb.SearchMovies(query, page, language, region, year)
}

func (s *MovieService) GetByID(id int, language string) (*models.Movie, error) {
	return s.tmdb.GetMovie(id, language)
}

func (s *MovieService) GetPopular(page int, language, region string) (*models.TMDBResponse, error) {
	return s.tmdb.GetPopularMovies(page, language, region)
}

func (s *MovieService) GetTopRated(page int, language, region string) (*models.TMDBResponse, error) {
	return s.tmdb.GetTopRatedMovies(page, language, region)
}

func (s *MovieService) GetUpcoming(page int, language, region string) (*models.TMDBResponse, error) {
	return s.tmdb.GetUpcomingMovies(page, language, region)
}

func (s *MovieService) GetNowPlaying(page int, language, region string) (*models.TMDBResponse, error) {
	return s.tmdb.GetNowPlayingMovies(page, language, region)
}

func (s *MovieService) GetRecommendations(id, page int, language string) (*models.TMDBResponse, error) {
	return s.tmdb.GetMovieRecommendations(id, page, language)
}

func (s *MovieService) GetSimilar(id, page int, language string) (*models.TMDBResponse, error) {
	return s.tmdb.GetSimilarMovies(id, page, language)
}



func (s *MovieService) GetExternalIDs(id int) (*models.ExternalIDs, error) {
	return s.tmdb.GetMovieExternalIDs(id)
}