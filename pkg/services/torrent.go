package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"neomovies-api/pkg/models"
)

type TorrentService struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewTorrentServiceWithConfig(baseURL, apiKey string) *TorrentService {
	return &TorrentService{
		client:  &http.Client{Timeout: 8 * time.Second},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func NewTorrentService() *TorrentService {
	return &TorrentService{
		client:  &http.Client{Timeout: 8 * time.Second},
		baseURL: "http://redapi.cfhttp.top",
		apiKey:  "",
	}
}

// SearchTorrents - основной метод поиска торрентов через RedAPI
func (s *TorrentService) SearchTorrents(params map[string]string) (*models.TorrentSearchResponse, error) {
	searchParams := url.Values{}

	for key, value := range params {
		if value != "" {
			if key == "category" {
				searchParams.Add("category[]", value)
			} else {
				searchParams.Add(key, value)
			}
		}
	}

	if s.apiKey != "" {
		searchParams.Add("apikey", s.apiKey)
	}

	searchURL := fmt.Sprintf("%s/api/v2.0/indexers/all/results?%s", s.baseURL, searchParams.Encode())

	resp, err := s.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search torrents: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var redAPIResponse models.RedAPIResponse
	if err := json.Unmarshal(body, &redAPIResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := s.parseRedAPIResults(redAPIResponse)

	return &models.TorrentSearchResponse{
		Query:   params["query"],
		Results: results,
		Total:   len(results),
	}, nil
}

// parseRedAPIResults преобразует результаты RedAPI в наш формат
func (s *TorrentService) parseRedAPIResults(data models.RedAPIResponse) []models.TorrentResult {
	var results []models.TorrentResult

	for _, torrent := range data.Results {
		var sizeStr string
		switch v := torrent.Size.(type) {
		case string:
			sizeStr = v
		case float64:
			sizeStr = fmt.Sprintf("%.0f", v)
		case int:
			sizeStr = fmt.Sprintf("%d", v)
		default:
			sizeStr = ""
		}

		result := models.TorrentResult{
			Title:       torrent.Title,
			Tracker:     torrent.Tracker,
			Size:        sizeStr,
			Seeders:     torrent.Seeders,
			Peers:       torrent.Peers,
			MagnetLink:  torrent.MagnetUri,
			PublishDate: torrent.PublishDate,
			Category:    torrent.CategoryDesc,
			Details:     torrent.Details,
			Source:      "RedAPI",
		}

		if torrent.Info != nil {
			switch v := torrent.Info.Quality.(type) {
			case string:
				result.Quality = v
			case float64:
				result.Quality = fmt.Sprintf("%.0fp", v)
			case int:
				result.Quality = fmt.Sprintf("%dp", v)
			}

			result.Voice = torrent.Info.Voices
			result.Types = torrent.Info.Types
			result.Seasons = torrent.Info.Seasons
		}

		if result.Quality == "" {
			result.Quality = s.ExtractQuality(result.Title)
		}

		results = append(results, result)
	}

	return results
}

// SearchTorrentsByIMDbID - поиск по IMDB ID с поддержкой всех функций
func (s *TorrentService) SearchTorrentsByIMDbID(tmdbService *TMDBService, imdbID, mediaType string, options *models.TorrentSearchOptions) (*models.TorrentSearchResponse, error) {
	title, originalTitle, year, err := s.getTitleFromTMDB(tmdbService, imdbID, mediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to get title from TMDB: %w", err)
	}

	params := map[string]string{
		"imdb":           imdbID,
		"query":          title,
		"title_original": originalTitle,
		"year":           year,
	}

	switch mediaType {
	case "movie":
		params["is_serial"] = "1"
		params["category"] = "2000"
	case "serial", "series", "tv":
		params["is_serial"] = "2"
		params["category"] = "5000"
	case "anime":
		params["is_serial"] = "5"
		params["category"] = "5070"
	}

	if options != nil && options.Season != nil && *options.Season > 0 {
		params["season"] = strconv.Itoa(*options.Season)
	}

	response, err := s.SearchTorrents(params)
	if err != nil {
		return nil, err
	}

	if options != nil {
		response.Results = s.FilterByContentType(response.Results, options.ContentType)
		response.Results = s.FilterTorrents(response.Results, options)
		response.Results = s.sortTorrents(response.Results, options.SortBy, options.SortOrder)
	}
	response.Total = len(response.Results)

	if len(response.Results) < 5 && (mediaType == "serial" || mediaType == "series" || mediaType == "tv") && options != nil && options.Season != nil {
		paramsNoSeason := map[string]string{
			"imdb":           imdbID,
			"query":          title,
			"title_original": originalTitle,
			"year":           year,
			"is_serial":      "2",
			"category":       "5000",
		}
		fallbackResp, err := s.SearchTorrents(paramsNoSeason)
		if err == nil {
			filtered := s.filterBySeason(fallbackResp.Results, *options.Season)
			all := append(response.Results, filtered...)
			unique := make([]models.TorrentResult, 0, len(all))
			seen := make(map[string]bool)
			for _, t := range all {
				if !seen[t.MagnetLink] {
					unique = append(unique, t)
					seen[t.MagnetLink] = true
				}
			}
			response.Results = unique
		}
	}

	return response, nil
}


// SearchMovies - поиск фильмов с дополнительной фильтрацией
func (s *TorrentService) SearchMovies(title, originalTitle, year string) (*models.TorrentSearchResponse, error) {
	params := map[string]string{
		"title":          title,
		"title_original": originalTitle,
		"year":           year,
		"is_serial":      "1",
		"category":       "2000",
	}

	response, err := s.SearchTorrents(params)
	if err != nil {
		return nil, err
	}

	response.Results = s.FilterByContentType(response.Results, "movie")
	response.Total = len(response.Results)

	return response, nil
}

// SearchSeries - поиск сериалов с поддержкой fallback и фильтрации по сезону
func (s *TorrentService) SearchSeries(title, originalTitle, year string, season *int) (*models.TorrentSearchResponse, error) {
	params := map[string]string{
		"title":          title,
		"title_original": originalTitle,
		"year":           year,
		"is_serial":      "2",
		"category":       "5000",
	}
	if season != nil {
		params["season"] = strconv.Itoa(*season)
	}

	response, err := s.SearchTorrents(params)
	if err != nil {
		return nil, err
	}

	// Если указан сезон и результатов мало, делаем fallback-поиск без сезона и фильтруем на клиенте
	if season != nil && len(response.Results) < 5 {
		paramsNoSeason := map[string]string{
			"title":          title,
			"title_original": originalTitle,
			"year":           year,
			"is_serial":      "2",
			"category":       "5000",
		}
		fallbackResp, err := s.SearchTorrents(paramsNoSeason)
		if err == nil {
			filtered := s.filterBySeason(fallbackResp.Results, *season)
			// Объединяем и убираем дубликаты по MagnetLink
			all := append(response.Results, filtered...)
			unique := make([]models.TorrentResult, 0, len(all))
			seen := make(map[string]bool)
			for _, t := range all {
				if !seen[t.MagnetLink] {
					unique = append(unique, t)
					seen[t.MagnetLink] = true
				}
			}
			response.Results = unique
		}
	}

	response.Results = s.FilterByContentType(response.Results, "serial")
	response.Total = len(response.Results)
	return response, nil
}

// filterBySeason - фильтрация результатов по сезону (аналогично JS)
func (s *TorrentService) filterBySeason(results []models.TorrentResult, season int) []models.TorrentResult {
	if season == 0 {
		return results
	}
	filtered := make([]models.TorrentResult, 0, len(results))
	seasonRegex := regexp.MustCompile(`(?i)(?:s|сезон)[\s:]*(\d+)|(\d+)\s*сезон`)
	for _, torrent := range results {
		found := false
		// Проверяем поле seasons
		for _, s := range torrent.Seasons {
			if s == season {
				found = true
				break
			}
		}
		if found {
			filtered = append(filtered, torrent)
			continue
		}
		// Проверяем в названии
		matches := seasonRegex.FindAllStringSubmatch(torrent.Title, -1)
		for _, match := range matches {
			seasonNumber := 0
			if match[1] != "" {
				seasonNumber, _ = strconv.Atoi(match[1])
			} else if match[2] != "" {
				seasonNumber, _ = strconv.Atoi(match[2])
			}
			if seasonNumber == season {
				filtered = append(filtered, torrent)
				break
			}
		}
	}
	return filtered
}

// SearchAnime - поиск аниме
func (s *TorrentService) SearchAnime(title, originalTitle, year string) (*models.TorrentSearchResponse, error) {
	params := map[string]string{
		"title":          title,
		"title_original": originalTitle,
		"year":           year,
		"is_serial":      "5",
		"category":       "5070",
	}

	response, err := s.SearchTorrents(params)
	if err != nil {
		return nil, err
	}

	response.Results = s.FilterByContentType(response.Results, "anime")
	response.Total = len(response.Results)

	return response, nil
}

// AllohaResponse - структура ответа от Alloha API
type AllohaResponse struct {
	Status string `json:"status"`
	Data   struct {
		Name         string `json:"name"`
		OriginalName string `json:"original_name"`
		Year         int    `json:"year"`
		Category     int    `json:"category"` // 1-фильм, 2-сериал
	} `json:"data"`
}

// getMovieInfoByIMDB - получение информации через Alloha API (как в JavaScript версии)
func (s *TorrentService) getMovieInfoByIMDB(imdbID string) (string, string, string, error) {
	// Используем тот же токен что и в JavaScript версии
	endpoint := fmt.Sprintf("https://api.alloha.tv/?token=04941a9a3ca3ac16e2b4327347bbc1&imdb=%s", imdbID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", "", "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	var allohaResponse AllohaResponse
	if err := json.Unmarshal(body, &allohaResponse); err != nil {
		return "", "", "", err
	}

	if allohaResponse.Status != "success" {
		return "", "", "", fmt.Errorf("no results found for IMDB ID: %s", imdbID)
	}

	title := allohaResponse.Data.Name
	originalTitle := allohaResponse.Data.OriginalName
	year := ""
	if allohaResponse.Data.Year > 0 {
		year = strconv.Itoa(allohaResponse.Data.Year)
	}

	return title, originalTitle, year, nil
}

// getTitleFromTMDB - получение информации из TMDB (с fallback на Alloha API)
func (s *TorrentService) getTitleFromTMDB(tmdbService *TMDBService, imdbID, mediaType string) (string, string, string, error) {
	// Сначала пробуем Alloha API (как в JavaScript версии)
	title, originalTitle, year, err := s.getMovieInfoByIMDB(imdbID)
	if err == nil {
		return title, originalTitle, year, nil
	}

	// Если Alloha API не работает, пробуем TMDB API
	endpoint := fmt.Sprintf("https://api.themoviedb.org/3/find/%s", imdbID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", "", "", err
	}

	params := url.Values{}
	params.Set("external_source", "imdb_id")
	params.Set("language", "ru-RU")
	req.URL.RawQuery = params.Encode()

	req.Header.Set("Authorization", "Bearer "+tmdbService.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	var findResponse struct {
		MovieResults []struct {
			Title         string `json:"title"`
			OriginalTitle string `json:"original_title"`
			ReleaseDate   string `json:"release_date"`
		} `json:"movie_results"`
		TVResults []struct {
			Name         string `json:"name"`
			OriginalName string `json:"original_name"`
			FirstAirDate string `json:"first_air_date"`
		} `json:"tv_results"`
	}

	if err := json.Unmarshal(body, &findResponse); err != nil {
		return "", "", "", err
	}

	if mediaType == "movie" && len(findResponse.MovieResults) > 0 {
		movie := findResponse.MovieResults[0]
		title := movie.Title
		originalTitle := movie.OriginalTitle
		year := ""
		if movie.ReleaseDate != "" {
			year = movie.ReleaseDate[:4]
		}
		return title, originalTitle, year, nil
	}

	if (mediaType == "tv" || mediaType == "series") && len(findResponse.TVResults) > 0 {
		tv := findResponse.TVResults[0]
		title := tv.Name
		originalTitle := tv.OriginalName
		year := ""
		if tv.FirstAirDate != "" {
			year = tv.FirstAirDate[:4]
		}
		return title, originalTitle, year, nil
	}

	return "", "", "", fmt.Errorf("no results found for IMDB ID: %s", imdbID)
}

// FilterByContentType - фильтрация по типу контента (как в JS)
func (s *TorrentService) FilterByContentType(results []models.TorrentResult, contentType string) []models.TorrentResult {
	if contentType == "" {
		return results
	}
	var filtered []models.TorrentResult
	for _, torrent := range results {
		// Фильтрация по полю types, если оно есть
		if len(torrent.Types) > 0 {
			switch contentType {
			case "movie":
				if s.containsAny(torrent.Types, []string{"movie", "multfilm", "documovie"}) {
					filtered = append(filtered, torrent)
				}
			case "serial", "series", "tv":
				if s.containsAny(torrent.Types, []string{"serial", "multserial", "docuserial", "tvshow"}) {
					filtered = append(filtered, torrent)
				}
			case "anime":
				if s.contains(torrent.Types, "anime") {
					filtered = append(filtered, torrent)
				}
			}
			continue
		}
		// Фильтрация по названию, если types недоступно
		title := strings.ToLower(torrent.Title)
		switch contentType {
		case "movie":
			if !regexp.MustCompile(`(?i)(сезон|серии|series|season|эпизод)`).MatchString(title) {
				filtered = append(filtered, torrent)
			}
		case "serial", "series", "tv":
			if regexp.MustCompile(`(?i)(сезон|серии|series|season|эпизод)`).MatchString(title) {
				filtered = append(filtered, torrent)
			}
		case "anime":
			if torrent.Category == "TV/Anime" || regexp.MustCompile(`(?i)anime`).MatchString(title) {
				filtered = append(filtered, torrent)
			}
		default:
			filtered = append(filtered, torrent)
		}
	}
	return filtered
}

// FilterTorrents - фильтрация торрентов по опциям
func (s *TorrentService) FilterTorrents(torrents []models.TorrentResult, options *models.TorrentSearchOptions) []models.TorrentResult {
	if options == nil {
		return torrents
	}

	var filtered []models.TorrentResult

	for _, torrent := range torrents {
		// Фильтрация по качеству
		if len(options.Quality) > 0 {
			found := false
			for _, quality := range options.Quality {
				if strings.EqualFold(torrent.Quality, quality) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Фильтрация по минимальному качеству
		if options.MinQuality != "" && !s.qualityMeetsMinimum(torrent.Quality, options.MinQuality) {
			continue
		}

		// Фильтрация по максимальному качеству
		if options.MaxQuality != "" && !s.qualityMeetsMaximum(torrent.Quality, options.MaxQuality) {
			continue
		}

		// Исключение качеств
		if len(options.ExcludeQualities) > 0 {
			excluded := false
			for _, excludeQuality := range options.ExcludeQualities {
				if strings.EqualFold(torrent.Quality, excludeQuality) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		// Фильтрация по HDR
		if options.HDR != nil {
			hasHDR := regexp.MustCompile(`(?i)(hdr|dolby.vision|dv)`).MatchString(torrent.Title)
			if *options.HDR != hasHDR {
				continue
			}
		}

		// Фильтрация по HEVC
		if options.HEVC != nil {
			hasHEVC := regexp.MustCompile(`(?i)(hevc|h\.265|x265)`).MatchString(torrent.Title)
			if *options.HEVC != hasHEVC {
				continue
			}
		}

		// Фильтрация по сезону (дополнительная на клиенте)
		if options.Season != nil {
			if !s.matchesSeason(torrent, *options.Season) {
				continue
			}
		}

		filtered = append(filtered, torrent)
	}

	return filtered
}

// matchesSeason - проверка соответствия сезону
func (s *TorrentService) matchesSeason(torrent models.TorrentResult, season int) bool {
	// Проверяем в поле seasons
	for _, s := range torrent.Seasons {
		if s == season {
			return true
		}
	}

	// Проверяем в названии
	seasonRegex := regexp.MustCompile(`(?i)(?:s|сезон)[\s:]*(\d+)|(\d+)\s*сезон`)
	matches := seasonRegex.FindAllStringSubmatch(torrent.Title, -1)
	for _, match := range matches {
		seasonNumber := 0
		if match[1] != "" {
			seasonNumber, _ = strconv.Atoi(match[1])
		} else if match[2] != "" {
			seasonNumber, _ = strconv.Atoi(match[2])
		}
		if seasonNumber == season {
			return true
		}
	}

	return false
}

// ExtractQuality - извлечение качества из названия
func (s *TorrentService) ExtractQuality(title string) string {
	title = strings.ToUpper(title)

	qualityPatterns := []struct {
		pattern string
		quality string
	}{
		{`2160P|4K`, "2160p"},
		{`1440P`, "1440p"},
		{`1080P`, "1080p"},
		{`720P`, "720p"},
		{`480P`, "480p"},
		{`360P`, "360p"},
	}

	for _, qp := range qualityPatterns {
		if matched, _ := regexp.MatchString(qp.pattern, title); matched {
			if qp.quality == "2160p" {
				return "4K"
			}
			return qp.quality
		}
	}

	return "Unknown"
}

// sortTorrents - сортировка результатов
func (s *TorrentService) sortTorrents(torrents []models.TorrentResult, sortBy, sortOrder string) []models.TorrentResult {
	if sortBy == "" {
		sortBy = "seeders"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	sort.Slice(torrents, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "seeders":
			less = torrents[i].Seeders < torrents[j].Seeders
		case "size":
			less = s.compareSizes(torrents[i].Size, torrents[j].Size)
		case "date":
			t1, _ := time.Parse(time.RFC3339, torrents[i].PublishDate)
			t2, _ := time.Parse(time.RFC3339, torrents[j].PublishDate)
			less = t1.Before(t2)
		default:
			less = torrents[i].Seeders < torrents[j].Seeders
		}

		if sortOrder == "asc" {
			return less
		}
		return !less
	})

	return torrents
}

// GroupByQuality - группировка по качеству
func (s *TorrentService) GroupByQuality(results []models.TorrentResult) map[string][]models.TorrentResult {
	groups := make(map[string][]models.TorrentResult)

	for _, torrent := range results {
		quality := torrent.Quality
		if quality == "" {
			quality = "unknown"
		}

		// Объединяем 4K и 2160p в одну группу
		if quality == "2160p" {
			quality = "4K"
		}

		groups[quality] = append(groups[quality], torrent)
	}

	// Сортируем торренты внутри каждой группы по сидам
	for quality := range groups {
		sort.Slice(groups[quality], func(i, j int) bool {
			return groups[quality][i].Seeders > groups[quality][j].Seeders
		})
	}

	return groups
}

// GroupBySeason - группировка по сезонам
func (s *TorrentService) GroupBySeason(results []models.TorrentResult) map[string][]models.TorrentResult {
	groups := make(map[string][]models.TorrentResult)

	for _, torrent := range results {
		seasons := make(map[int]bool)

		// Извлекаем сезоны из поля seasons
		for _, season := range torrent.Seasons {
			seasons[season] = true
		}

		// Извлекаем сезоны из названия
		seasonRegex := regexp.MustCompile(`(?i)(?:s|сезон)[\s:]*(\d+)|(\d+)\s*сезон`)
		matches := seasonRegex.FindAllStringSubmatch(torrent.Title, -1)
		for _, match := range matches {
			seasonNumber := 0
			if match[1] != "" {
				seasonNumber, _ = strconv.Atoi(match[1])
			} else if match[2] != "" {
				seasonNumber, _ = strconv.Atoi(match[2])
			}
			if seasonNumber > 0 {
				seasons[seasonNumber] = true
			}
		}

		// Если сезоны не найдены, добавляем в группу "unknown"
		if len(seasons) == 0 {
			groups["Неизвестно"] = append(groups["Неизвестно"], torrent)
		} else {
			// Добавляем торрент во все соответствующие группы сезонов
			for season := range seasons {
				seasonKey := fmt.Sprintf("Сезон %d", season)
				// Проверяем дубликаты
				found := false
				for _, existing := range groups[seasonKey] {
					if existing.MagnetLink == torrent.MagnetLink {
						found = true
						break
					}
				}
				if !found {
					groups[seasonKey] = append(groups[seasonKey], torrent)
				}
			}
		}
	}

	// Сортируем торренты внутри каждой группы по сидам
	for season := range groups {
		sort.Slice(groups[season], func(i, j int) bool {
			return groups[season][i].Seeders > groups[season][j].Seeders
		})
	}

	return groups
}

// GetAvailableSeasons - получение доступных сезонов для сериала
func (s *TorrentService) GetAvailableSeasons(title, originalTitle, year string) ([]int, error) {
	response, err := s.SearchSeries(title, originalTitle, year, nil)
	if err != nil {
		return nil, err
	}

	seasonsSet := make(map[int]bool)

	for _, torrent := range response.Results {
		// Извлекаем из поля seasons
		for _, season := range torrent.Seasons {
			seasonsSet[season] = true
		}

		// Извлекаем из названия
		seasonRegex := regexp.MustCompile(`(?i)(?:s|сезон)[\s:]*(\d+)|(\d+)\s*сезон`)
		matches := seasonRegex.FindAllStringSubmatch(torrent.Title, -1)
		for _, match := range matches {
			seasonNumber := 0
			if match[1] != "" {
				seasonNumber, _ = strconv.Atoi(match[1])
			} else if match[2] != "" {
				seasonNumber, _ = strconv.Atoi(match[2])
			}
			if seasonNumber > 0 {
				seasonsSet[seasonNumber] = true
			}
		}
	}

	var seasons []int
	for season := range seasonsSet {
		seasons = append(seasons, season)
	}

	sort.Ints(seasons)
	return seasons, nil
}

// SearchByImdb - поиск по IMDB ID (movie/serial/anime).
func (s *TorrentService) SearchByImdb(imdbID, contentType string, season *int) ([]models.TorrentResult, error) {
	if imdbID == "" || !strings.HasPrefix(imdbID, "tt") {
		return nil, fmt.Errorf("Неверный формат IMDB ID. Должен быть в формате tt1234567")
	}

	// НЕ добавляем title, originalTitle, year, чтобы запрос не был слишком строгим.
	params := map[string]string{
		"imdb": imdbID,
	}

	// Определяем тип контента для API
	switch contentType {
	case "movie":
		params["is_serial"] = "1"
		params["category"] = "2000"
	case "serial", "series", "tv":
		params["is_serial"] = "2"
		params["category"] = "5000"
	case "anime":
		params["is_serial"] = "5"
		params["category"] = "5070"
	default:
		// Значение по умолчанию на случай неизвестного типа
		params["is_serial"] = "1"
		params["category"] = "2000"
	}

	// Параметр season можно оставить, он полезен
	if season != nil && *season > 0 {
		params["season"] = strconv.Itoa(*season)
	}

	resp, err := s.SearchTorrents(params)
	if err != nil {
		return nil, err
	}
	results := resp.Results

	// Fallback для сериалов: если указан сезон и результатов мало, ищем без сезона и фильтруем на клиенте
	if s.contains([]string{"serial", "series", "tv"}, contentType) && season != nil && len(results) < 5 {
		paramsNoSeason := map[string]string{
			"imdb":      imdbID,
			"is_serial": "2",
			"category":  "5000",
		}
		
		fallbackResp, err := s.SearchTorrents(paramsNoSeason)
		if err == nil {
			filtered := s.filterBySeason(fallbackResp.Results, *season)
			// Объединяем и убираем дубликаты по MagnetLink
			all := append(results, filtered...)
			unique := make([]models.TorrentResult, 0, len(all))
			seen := make(map[string]bool)
			for _, t := range all {
				if !seen[t.MagnetLink] {
					unique = append(unique, t)
					seen[t.MagnetLink] = true
				}
			}
			results = unique
		}
	}

	// Финальная фильтрация по типу контента на стороне клиента для надежности
	results = s.FilterByContentType(results, contentType)
	return results, nil
}

// ############# ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ #############

func (s *TorrentService) qualityMeetsMinimum(quality, minQuality string) bool {
	qualityOrder := map[string]int{
		"360p": 1, "480p": 2, "720p": 3, "1080p": 4, "1440p": 5, "4K": 6, "2160p": 6,
	}

	currentLevel, ok1 := qualityOrder[strings.ToLower(quality)]
	minLevel, ok2 := qualityOrder[strings.ToLower(minQuality)]

	if !ok1 || !ok2 {
		return true // Если качество не определено, не фильтруем
	}

	return currentLevel >= minLevel
}

func (s *TorrentService) qualityMeetsMaximum(quality, maxQuality string) bool {
	qualityOrder := map[string]int{
		"360p": 1, "480p": 2, "720p": 3, "1080p": 4, "1440p": 5, "4K": 6, "2160p": 6,
	}

	currentLevel, ok1 := qualityOrder[strings.ToLower(quality)]
	maxLevel, ok2 := qualityOrder[strings.ToLower(maxQuality)]

	if !ok1 || !ok2 {
		return true // Если качество не определено, не фильтруем
	}

	return currentLevel <= maxLevel
}

func (s *TorrentService) parseSize(sizeStr string) int64 {
	val, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func (s *TorrentService) compareSizes(size1, size2 string) bool {
	return s.parseSize(size1) < s.parseSize(size2)
}

func (s *TorrentService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func (s *TorrentService) containsAny(slice []string, items []string) bool {
	for _, item := range items {
		if s.contains(slice, item) {
			return true
		}
	}
	return false
}
