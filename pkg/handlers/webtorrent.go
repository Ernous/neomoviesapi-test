package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"neomovies-api/pkg/models"
	"neomovies-api/pkg/services"
)

type WebTorrentHandler struct {
	tmdbService *services.TMDBService
}

func NewWebTorrentHandler(tmdbService *services.TMDBService) *WebTorrentHandler {
	return &WebTorrentHandler{
		tmdbService: tmdbService,
	}
}

// Структура для ответа с метаданными
type MediaMetadata struct {
	ID           int               `json:"id"`
	Title        string            `json:"title"`
	Type         string            `json:"type"` // "movie" or "tv"
	Year         int               `json:"year,omitempty"`
	PosterPath   string            `json:"posterPath,omitempty"`
	BackdropPath string            `json:"backdropPath,omitempty"`
	Overview     string            `json:"overview,omitempty"`
	Seasons      []SeasonMetadata  `json:"seasons,omitempty"`
	Episodes     []EpisodeMetadata `json:"episodes,omitempty"`
	Runtime      int               `json:"runtime,omitempty"`
	Genres       []models.Genre    `json:"genres,omitempty"`
}

type SeasonMetadata struct {
	SeasonNumber int               `json:"seasonNumber"`
	Name         string            `json:"name"`
	Episodes     []EpisodeMetadata `json:"episodes"`
}

type EpisodeMetadata struct {
	EpisodeNumber int    `json:"episodeNumber"`
	SeasonNumber  int    `json:"seasonNumber"`
	Name          string `json:"name"`
	Overview      string `json:"overview,omitempty"`
	Runtime       int    `json:"runtime,omitempty"`
	StillPath     string `json:"stillPath,omitempty"`
}

// Открытие плеера с магнет ссылкой
func (h *WebTorrentHandler) OpenPlayer(w http.ResponseWriter, r *http.Request) {
	magnetLink := r.Header.Get("X-Magnet-Link")
	if magnetLink == "" {
		magnetLink = r.URL.Query().Get("magnet")
	}
	
	if magnetLink == "" {
		http.Error(w, "Magnet link is required", http.StatusBadRequest)
		return
	}

	// Декодируем magnet ссылку если она закодирована
	decodedMagnet, err := url.QueryUnescape(magnetLink)
	if err != nil {
		decodedMagnet = magnetLink
	}

	// Отдаем HTML страницу с плеером
	tmpl := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NeoMovies WebTorrent Player</title>
    <script src="https://cdn.jsdelivr.net/npm/webtorrent@latest/webtorrent.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            background: #000;
            color: #fff;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            overflow: hidden;
        }
        
        .player-container {
            position: relative;
            width: 100vw;
            height: 100vh;
            display: flex;
            flex-direction: column;
        }
        
        .loading {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            text-align: center;
            z-index: 100;
        }
        
        .loading-spinner {
            border: 4px solid #333;
            border-top: 4px solid #fff;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 20px;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        
        .media-info {
            position: absolute;
            top: 20px;
            left: 20px;
            z-index: 50;
            background: rgba(0,0,0,0.8);
            padding: 15px;
            border-radius: 8px;
            max-width: 400px;
            display: none;
        }
        
        .media-title {
            font-size: 18px;
            font-weight: bold;
            margin-bottom: 5px;
        }
        
        .media-overview {
            font-size: 14px;
            color: #ccc;
            line-height: 1.4;
        }
        
        .controls {
            position: absolute;
            bottom: 20px;
            left: 20px;
            right: 20px;
            z-index: 50;
            background: rgba(0,0,0,0.8);
            padding: 15px;
            border-radius: 8px;
            display: none;
        }
        
        .file-list {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            margin-bottom: 15px;
        }
        
        .file-item {
            background: #333;
            border: none;
            color: #fff;
            padding: 8px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            transition: background 0.2s;
        }
        
        .file-item:hover {
            background: #555;
        }
        
        .file-item.active {
            background: #007bff;
        }
        
        .episode-info {
            font-size: 14px;
            margin-bottom: 10px;
            color: #ccc;
        }
        
        video {
            width: 100%;
            height: 100%;
            object-fit: contain;
        }
        
        .error {
            color: #ff4444;
            text-align: center;
            padding: 20px;
        }
    </style>
</head>
<body>
    <div class="player-container">
        <div class="loading" id="loading">
            <div class="loading-spinner"></div>
            <div>Загружаем торрент...</div>
            <div id="loadingProgress" style="margin-top: 10px; font-size: 12px;"></div>
        </div>
        
        <div class="media-info" id="mediaInfo">
            <div class="media-title" id="mediaTitle"></div>
            <div class="media-overview" id="mediaOverview"></div>
        </div>
        
        <div class="controls" id="controls">
            <div class="episode-info" id="episodeInfo"></div>
            <div class="file-list" id="fileList"></div>
        </div>
        
        <video id="videoPlayer" controls style="display: none;"></video>
    </div>

    <script>
        const magnetLink = {{.MagnetLink}};
        const baseURL = {{.BaseURL}};
        const client = new WebTorrent();
        
        let currentTorrent = null;
        let mediaMetadata = null;
        
        const elements = {
            loading: document.getElementById('loading'),
            mediaInfo: document.getElementById('mediaInfo'),
            mediaTitle: document.getElementById('mediaTitle'),
            mediaOverview: document.getElementById('mediaOverview'),
            controls: document.getElementById('controls'),
            episodeInfo: document.getElementById('episodeInfo'),
            fileList: document.getElementById('fileList'),
            videoPlayer: document.getElementById('videoPlayer'),
            loadingProgress: document.getElementById('loadingProgress')
        };
        
        // Загружаем торрент
        client.add(magnetLink, onTorrent);
        
        function onTorrent(torrent) {
            currentTorrent = torrent;
            console.log('Торрент загружен:', torrent.name);
            
            // Получаем метаданные через API
            fetchMediaMetadata(torrent.name);
            
            // Фильтруем только видео файлы
            const videoFiles = torrent.files.filter(file => 
                /\.(mp4|avi|mkv|mov|wmv|flv|webm|m4v)$/i.test(file.name)
            );
            
            if (videoFiles.length === 0) {
                showError('Видео файлы не найдены в торренте');
                return;
            }
            
            // Показываем список файлов
            renderFileList(videoFiles);
            
            // Автоматически выбираем первый файл
            if (videoFiles.length > 0) {
                playFile(videoFiles[0], 0);
            }
            
            elements.loading.style.display = 'none';
            elements.controls.style.display = 'block';
        }
        
        function fetchMediaMetadata(torrentName) {
            // Извлекаем название для поиска из имени торрента
            const searchQuery = extractTitleFromTorrentName(torrentName);
            
            fetch(baseURL + '/api/v1/webtorrent/metadata?query=' + encodeURIComponent(searchQuery))
                .then(response => response.json())
                .then(data => {
                    if (data.success && data.data) {
                        mediaMetadata = data.data;
                        displayMediaInfo(mediaMetadata);
                    }
                })
                .catch(error => console.log('Метаданные не найдены:', error));
        }
        
        function extractTitleFromTorrentName(name) {
            // Убираем расширения файлов, качество, кодеки и т.д.
            let title = name
                .replace(/\.(mp4|avi|mkv|mov|wmv|flv|webm|m4v)$/i, '')
                .replace(/\b(1080p|720p|480p|4K|BluRay|WEBRip|DVDRip|HDTV|x264|x265|HEVC|DTS|AC3)\b/gi, '')
                .replace(/\b(S\d{1,2}E\d{1,2}|\d{4})\b/g, '')
                .replace(/[\.\-_\[\]()]/g, ' ')
                .replace(/\s+/g, ' ')
                .trim();
            
            return title;
        }
        
        function displayMediaInfo(metadata) {
            elements.mediaTitle.textContent = metadata.title + (metadata.year ? ' (' + metadata.year + ')' : '');
            elements.mediaOverview.textContent = metadata.overview || '';
            elements.mediaInfo.style.display = 'block';
        }
        
        function renderFileList(files) {
            elements.fileList.innerHTML = '';
            
            files.forEach((file, index) => {
                const button = document.createElement('button');
                button.className = 'file-item';
                button.textContent = getDisplayName(file.name, index);
                button.onclick = () => playFile(file, index);
                elements.fileList.appendChild(button);
            });
        }
        
        function getDisplayName(fileName, index) {
            if (!mediaMetadata) {
                return fileName;
            }
            
            // Для сериалов пытаемся определить сезон и серию
            if (mediaMetadata.type === 'tv') {
                const episodeMatch = fileName.match(/S(\d{1,2})E(\d{1,2})/i);
                if (episodeMatch) {
                    const season = parseInt(episodeMatch[1]);
                    const episode = parseInt(episodeMatch[2]);
                    
                    const episodeData = mediaMetadata.episodes?.find(ep => 
                        ep.seasonNumber === season && ep.episodeNumber === episode
                    );
                    
                    if (episodeData) {
                        return 'S' + season + 'E' + episode + ': ' + episodeData.name;
                    }
                }
            }
            
            return mediaMetadata.title + ' - Файл ' + (index + 1);
        }
        
        function playFile(file, index) {
            // Убираем активный класс со всех кнопок
            document.querySelectorAll('.file-item').forEach(btn => btn.classList.remove('active'));
            // Добавляем активный класс к выбранной кнопке
            document.querySelectorAll('.file-item')[index].classList.add('active');
            
            // Обновляем информацию о серии
            updateEpisodeInfo(file.name, index);
            
            // Воспроизводим файл
            file.renderTo(elements.videoPlayer, (err) => {
                if (err) {
                    showError('Ошибка воспроизведения: ' + err.message);
                } else {
                    elements.videoPlayer.style.display = 'block';
                }
            });
        }
        
        function updateEpisodeInfo(fileName, index) {
            if (!mediaMetadata) {
                elements.episodeInfo.textContent = 'Файл: ' + fileName;
                return;
            }
            
            if (mediaMetadata.type === 'tv') {
                const episodeMatch = fileName.match(/S(\d{1,2})E(\d{1,2})/i);
                if (episodeMatch) {
                    const season = parseInt(episodeMatch[1]);
                    const episode = parseInt(episodeMatch[2]);
                    
                    const episodeData = mediaMetadata.episodes?.find(ep => 
                        ep.seasonNumber === season && ep.episodeNumber === episode
                    );
                    
                    if (episodeData) {
                        elements.episodeInfo.innerHTML = 
                            '<strong>Сезон ' + season + ', Серия ' + episode + '</strong><br>' +
                            episodeData.name + 
                            (episodeData.overview ? '<br><span style="color: #999; font-size: 12px;">' + episodeData.overview + '</span>' : '');
                        return;
                    }
                }
            }
            
            elements.episodeInfo.textContent = mediaMetadata.title + ' - Часть ' + (index + 1);
        }
        
        function showError(message) {
            elements.loading.innerHTML = '<div class="error">' + message + '</div>';
        }
        
        // Обработка прогресса загрузки
        client.on('torrent', (torrent) => {
            torrent.on('download', () => {
                const progress = Math.round(torrent.progress * 100);
                const downloadSpeed = (torrent.downloadSpeed / 1024 / 1024).toFixed(1);
                elements.loadingProgress.textContent = 
                    'Прогресс: ' + progress + '% | Скорость: ' + downloadSpeed + ' MB/s';
            });
        });
        
        // Глобальная обработка ошибок
        client.on('error', (err) => {
            showError('Ошибка торрент клиента: ' + err.message);
        });
        
        // Управление с клавиатуры
        document.addEventListener('keydown', (e) => {
            if (e.code === 'Space') {
                e.preventDefault();
                if (elements.videoPlayer.paused) {
                    elements.videoPlayer.play();
                } else {
                    elements.videoPlayer.pause();
                }
            }
        });
    </script>
</body>
</html>`

	// Создаем template и выполняем его
	t, err := template.New("player").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Определяем правильный baseURL
	baseURL := getSchemeForWebTorrent(r) + "://" + r.Host
	
	data := struct {
		MagnetLink string
		BaseURL    string
	}{
		MagnetLink: strconv.Quote(decodedMagnet),
		BaseURL:    strconv.Quote(baseURL),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Добавляем заголовки безопасности для HTTPS
	if getSchemeForWebTorrent(r) == "https" {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "upgrade-insecure-requests")
	}
	
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// API для получения метаданных фильма/сериала по названию
func (h *WebTorrentHandler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Пытаемся определить тип контента и найти его
	metadata, err := h.searchAndBuildMetadata(query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.APIResponse{
			Success: false,
			Message: "Media not found: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    metadata,
	})
}

func (h *WebTorrentHandler) searchAndBuildMetadata(query string) (*MediaMetadata, error) {
	// Сначала пробуем поиск по фильмам
	movieResults, err := h.tmdbService.SearchMovies(query, 1, "ru-RU", "", 0)
	if err == nil && len(movieResults.Results) > 0 {
		movie := movieResults.Results[0]
		
		// Получаем детальную информацию о фильме
		fullMovie, err := h.tmdbService.GetMovie(movie.ID, "ru-RU")
		if err == nil {
			return &MediaMetadata{
				ID:           fullMovie.ID,
				Title:        fullMovie.Title,
				Type:         "movie",
				Year:         extractYear(fullMovie.ReleaseDate),
				PosterPath:   fullMovie.PosterPath,
				BackdropPath: fullMovie.BackdropPath,
				Overview:     fullMovie.Overview,
				Runtime:      fullMovie.Runtime,
				Genres:       fullMovie.Genres,
			}, nil
		}
	}

	// Затем пробуем поиск по сериалам
	tvResults, err := h.tmdbService.SearchTV(query, 1, "ru-RU", 0)
	if err == nil && len(tvResults.Results) > 0 {
		tv := tvResults.Results[0]
		
		// Получаем детальную информацию о сериале
		fullTV, err := h.tmdbService.GetTVShow(tv.ID, "ru-RU")
		if err == nil {
			metadata := &MediaMetadata{
				ID:           fullTV.ID,
				Title:        fullTV.Name,
				Type:         "tv",
				Year:         extractYear(fullTV.FirstAirDate),
				PosterPath:   fullTV.PosterPath,
				BackdropPath: fullTV.BackdropPath,
				Overview:     fullTV.Overview,
				Genres:       fullTV.Genres,
			}

			// Получаем информацию о сезонах и сериях
			var allEpisodes []EpisodeMetadata
			for _, season := range fullTV.Seasons {
				if season.SeasonNumber == 0 {
					continue // Пропускаем спецвыпуски
				}

				seasonDetails, err := h.tmdbService.GetTVSeason(fullTV.ID, season.SeasonNumber, "ru-RU")
				if err == nil {
					var episodes []EpisodeMetadata
					for _, episode := range seasonDetails.Episodes {
						episodeData := EpisodeMetadata{
							EpisodeNumber: episode.EpisodeNumber,
							SeasonNumber:  season.SeasonNumber,
							Name:          episode.Name,
							Overview:      episode.Overview,
							Runtime:       episode.Runtime,
							StillPath:     episode.StillPath,
						}
						episodes = append(episodes, episodeData)
						allEpisodes = append(allEpisodes, episodeData)
					}

					metadata.Seasons = append(metadata.Seasons, SeasonMetadata{
						SeasonNumber: season.SeasonNumber,
						Name:         season.Name,
						Episodes:     episodes,
					})
				}
			}

			metadata.Episodes = allEpisodes
			return metadata, nil
		}
	}

	return nil, err
}

func extractYear(dateString string) int {
	if len(dateString) >= 4 {
		yearStr := dateString[:4]
		if year, err := strconv.Atoi(yearStr); err == nil {
			return year
		}
	}
	return 0
}

// getSchemeForWebTorrent определяет правильную схему протокола для WebTorrent
func getSchemeForWebTorrent(r *http.Request) string {
	// Проверяем заголовки прокси/балансировщика
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.ToLower(proto)
	}
	
	if proto := r.Header.Get("X-Forwarded-Protocol"); proto != "" {
		return strings.ToLower(proto)
	}
	
	if r.Header.Get("X-Forwarded-Ssl") == "on" {
		return "https"
	}
	
	if r.Header.Get("X-Url-Scheme") != "" {
		return strings.ToLower(r.Header.Get("X-Url-Scheme"))
	}
	
	// Cloudflare headers
	if r.Header.Get("CF-Visitor") != "" {
		if strings.Contains(r.Header.Get("CF-Visitor"), `"scheme":"https"`) {
			return "https"
		}
	}
	
	// Fallback к стандартной проверке TLS
	if r.TLS != nil {
		return "https"
	}
	
	return "http"
}

// Проверяем есть ли нужные методы в TMDB сервисе
func (h *WebTorrentHandler) checkMethods() {
	// Эти методы должны существовать в TMDBService:
	// - SearchMovies
	// - SearchTV  
	// - GetMovie
	// - GetTVShow
	// - GetTVSeason
}