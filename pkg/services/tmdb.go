package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"neomovies-api/pkg/models"
)

type TMDBService struct {
	accessToken string
	baseURL     string
	client      *http.Client
}

func NewTMDBService(accessToken string) *TMDBService {
	return &TMDBService{
		accessToken: accessToken,
		baseURL:     "https://api.themoviedb.org/3",
		client:      &http.Client{},
	}
}

func (s *TMDBService) makeRequest(endpoint string, target interface{}) error {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}

	// Используем Bearer токен вместо API key в query параметрах
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TMDB API error: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (s *TMDBService) SearchMovies(query string, page int, language, region string, year int) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("include_adult", "false")
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}
	
	if region != "" {
		params.Set("region", region)
	}
	
	if year > 0 {
		params.Set("year", strconv.Itoa(year))
	}

	endpoint := fmt.Sprintf("%s/search/movie?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) SearchMulti(query string, page int, language string) (*models.MultiSearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("include_adult", "false")

	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/search/multi?%s", s.baseURL, params.Encode())

	var response models.MultiSearchResponse
	err := s.makeRequest(endpoint, &response)
	if err != nil {
		return nil, err
	}

	// Фильтруем результаты: убираем "person", и без названия
	filteredResults := make([]models.MultiSearchResult, 0)
	for _, result := range response.Results {
		if result.MediaType == "person" {
			continue
		}

		hasTitle := false
		if result.MediaType == "movie" && result.Title != "" {
			hasTitle = true
		} else if result.MediaType == "tv" && result.Name != "" {
			hasTitle = true
		}

		if hasTitle {
			filteredResults = append(filteredResults, result)
		}
	}

	response.Results = filteredResults
	response.TotalResults = len(filteredResults)

	return &response, nil
}

func (s *TMDBService) SearchTVShows(query string, page int, language string, firstAirDateYear int) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("page", strconv.Itoa(page))
	params.Set("include_adult", "false")
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}
	
	if firstAirDateYear > 0 {
		params.Set("first_air_date_year", strconv.Itoa(firstAirDateYear))
	}

	endpoint := fmt.Sprintf("%s/search/tv?%s", s.baseURL, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetMovie(id int, language string) (*models.Movie, error) {
	params := url.Values{}
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/movie/%d?%s", s.baseURL, id, params.Encode())
	
	var movie models.Movie
	err := s.makeRequest(endpoint, &movie)
	return &movie, err
}

func (s *TMDBService) GetTVShow(id int, language string) (*models.TVShow, error) {
	params := url.Values{}
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/%d?%s", s.baseURL, id, params.Encode())
	
	var tvShow models.TVShow
	err := s.makeRequest(endpoint, &tvShow)
	return &tvShow, err
}

func (s *TMDBService) GetGenres(mediaType string, language string) (*models.GenresResponse, error) {
	params := url.Values{}
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/genre/%s/list?%s", s.baseURL, mediaType, params.Encode())
	
	var response models.GenresResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetAllGenres() (*models.GenresResponse, error) {
	// Получаем жанры фильмов
	movieGenres, err := s.GetGenres("movie", "ru-RU")
	if err != nil {
		return nil, err
	}

	// Получаем жанры сериалов
	tvGenres, err := s.GetGenres("tv", "ru-RU")
	if err != nil {
		return nil, err
	}

	// Объединяем жанры, убирая дубликаты
	allGenres := make(map[int]models.Genre)
	
	for _, genre := range movieGenres.Genres {
		allGenres[genre.ID] = genre
	}
	
	for _, genre := range tvGenres.Genres {
		allGenres[genre.ID] = genre
	}

	// Преобразуем обратно в слайс
	var genres []models.Genre
	for _, genre := range allGenres {
		genres = append(genres, genre)
	}

	return &models.GenresResponse{Genres: genres}, nil
}

func (s *TMDBService) GetPopularMovies(page int, language, region string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}
	
	if region != "" {
		params.Set("region", region)
	}

	endpoint := fmt.Sprintf("%s/movie/popular?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetTopRatedMovies(page int, language, region string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}
	
	if region != "" {
		params.Set("region", region)
	}

	endpoint := fmt.Sprintf("%s/movie/top_rated?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetUpcomingMovies(page int, language, region string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}
	
	if region != "" {
		params.Set("region", region)
	}

	endpoint := fmt.Sprintf("%s/movie/upcoming?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetNowPlayingMovies(page int, language, region string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}
	
	if region != "" {
		params.Set("region", region)
	}

	endpoint := fmt.Sprintf("%s/movie/now_playing?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetMovieRecommendations(id, page int, language string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/movie/%d/recommendations?%s", s.baseURL, id, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetSimilarMovies(id, page int, language string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/movie/%d/similar?%s", s.baseURL, id, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetPopularTVShows(page int, language string) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/popular?%s", s.baseURL, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetTopRatedTVShows(page int, language string) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/top_rated?%s", s.baseURL, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetOnTheAirTVShows(page int, language string) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/on_the_air?%s", s.baseURL, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetAiringTodayTVShows(page int, language string) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/airing_today?%s", s.baseURL, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetTVRecommendations(id, page int, language string) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/%d/recommendations?%s", s.baseURL, id, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetSimilarTVShows(id, page int, language string) (*models.TMDBTVResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/tv/%d/similar?%s", s.baseURL, id, params.Encode())
	
	var response models.TMDBTVResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) GetMovieExternalIDs(id int) (*models.ExternalIDs, error) {
	endpoint := fmt.Sprintf("%s/movie/%d/external_ids", s.baseURL, id)
	
	var ids models.ExternalIDs
	err := s.makeRequest(endpoint, &ids)
	return &ids, err
}

func (s *TMDBService) GetTVExternalIDs(id int) (*models.ExternalIDs, error) {
	endpoint := fmt.Sprintf("%s/tv/%d/external_ids", s.baseURL, id)
	
	var ids models.ExternalIDs
	err := s.makeRequest(endpoint, &ids)
	return &ids, err
}

func (s *TMDBService) DiscoverMoviesByGenre(genreID, page int, language string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("with_genres", strconv.Itoa(genreID))
	params.Set("sort_by", "popularity.desc")
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/discover/movie?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}

func (s *TMDBService) DiscoverTVByGenre(genreID, page int, language string) (*models.TMDBResponse, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("with_genres", strconv.Itoa(genreID))
	params.Set("sort_by", "popularity.desc")
	
	if language != "" {
		params.Set("language", language)
	} else {
		params.Set("language", "ru-RU")
	}

	endpoint := fmt.Sprintf("%s/discover/tv?%s", s.baseURL, params.Encode())
	
	var response models.TMDBResponse
	err := s.makeRequest(endpoint, &response)
	return &response, err
}