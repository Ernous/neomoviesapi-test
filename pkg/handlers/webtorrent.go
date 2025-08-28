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
	// Если magnet не передан, используем публичный тестовый торрент Sintel
	if magnetLink == "" {
		magnetLink = "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df6748d566095a10&dn=Sintel&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fsintel.torrent"
	}

	// Декодируем magnet ссылку если она закодирована
	decodedMagnet, err := url.QueryUnescape(magnetLink)
	if err != nil {
		decodedMagnet = magnetLink
	}

	// Проверяем, что это действительно magnet ссылка
	if !isValidMagnetLink(decodedMagnet) {
		http.Error(w, "Invalid magnet link format", http.StatusBadRequest)
		return
	}

	// Очищаем magnet ссылку от лишних символов
	cleanedMagnet := cleanMagnetLink(decodedMagnet)

	// Отдаем HTML страницу с плеером
	tmpl := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NeoMovies WebTorrent Player</title>
    <script src="https://cdn.jsdelivr.net/npm/webtorrent@latest/webtorrent.min.js"></script>
    <script src="https://unpkg.com/webtorrent@latest/webtorrent.min.js"></script>
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
        const identifier = {{.MagnetLinkJSON}};
        const client = new WebTorrent({
            tracker: {
                rtcConfig: {
                    iceServers: [
                        { urls: 'stun:stun.l.google.com:19302' },
                        { urls: 'stun:global.stun.twilio.com:3478?transport=udp' },
                        { urls: 'stun:stun.cloudflare.com:3478' }
                    ]
                }
            }
        });
        
        let currentTorrent = null;
        let mediaMetadata = null;
        let torrentTimeout = null;
        
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
        console.log('Попытка загрузки торрента/идентификатора:', identifier);
        
        // Проверяем, что WebTorrent доступен
        if (typeof WebTorrent === 'undefined') {
            showError('WebTorrent библиотека не загружена. Проверьте подключение к интернету.');
            return;
        }
        
        // Устанавливаем таймаут на 30 секунд
        torrentTimeout = setTimeout(() => {
            if (!currentTorrent) {
                showError('Таймаут загрузки торрента. Проверьте magnet ссылку и попробуйте снова.');
            }
        }, 30000);
        
        try {
            // Поддержка 40-символьного infoHash: собираем magnet автоматически
            let magnetLink = null;
            const hashLike = typeof identifier === 'string' && /^[a-fA-F0-9]{40}$/.test(identifier);
            if (typeof identifier === 'string' && identifier.startsWith('magnet:?')) {
                magnetLink = identifier;
            } else if (hashLike) {
                magnetLink = 'magnet:?xt=urn:btih:' + identifier;
            } else {
                // Пытаемся использовать как есть (magnet/http). В браузере http(s) .torrent может быть ограничен CORS
                magnetLink = identifier;
            }

            // Базовая валидация
            if (typeof magnetLink !== 'string' || !magnetLink) {
                showError('Пустой идентификатор торрента');
                return;
            }

            if (!magnetLink.startsWith('magnet:?')) {
                showError('Ожидалась magnet ссылка или 40-символьный hash');
                return;
            }

            if (!magnetLink.includes('xt=urn:btih:')) {
                showError('Magnet ссылка должна содержать info hash (xt=urn:btih:)');
                return;
            }

            // Список публичных WebRTC-трекеров для повышения доступности WebTorrent в браузере
            const addOptions = { maxWebConns: 8 };

            client.add(magnetLink, addOptions, (torrent) => {
                try {
                    onTorrent(torrent);
                } catch (err) {
                    console.error('Ошибка обработчика торрента:', err);
                    showError('Ошибка обработчика торрента: ' + err.message);
                }
            });
        } catch (error) {
            console.error('Ошибка при инициализации торрента:', error);
            showError('Ошибка инициализации: ' + error.message);
        }
        
        function onTorrent(torrent) {
            // Очищаем таймаут
            if (torrentTimeout) {
                clearTimeout(torrentTimeout);
                torrentTimeout = null;
            }
            
            currentTorrent = torrent;
            console.log('Торрент загружен:', torrent.name);
            
            // Получаем метаданные через API
            fetchMediaMetadata(torrent.name);
            
            // Приоритизируем видео файлы и игнорируем не-видео
            let videoFiles = torrent.files.filter(file => {
                const isVideo = /\.(mp4|avi|mkv|mov|wmv|flv|webm|m4v)$/i.test(file.name);
                if (isVideo && typeof file.select === 'function') {
                    file.select();
                }
                if (!isVideo && typeof file.deselect === 'function') {
                    file.deselect();
                }
                return isVideo;
            });
            // Сортируем по размеру по убыванию, чтобы выбирать основной файл
            try {
                // Сначала .mp4, затем по размеру по убыванию
                videoFiles.sort((a, b) => {
                    const aMp4 = /\.mp4$/i.test(a.name) ? 1 : 0;
                    const bMp4 = /\.mp4$/i.test(b.name) ? 1 : 0;
                    if (aMp4 !== bMp4) return bMp4 - aMp4;
                    return (b.length || 0) - (a.length || 0);
                });
            } catch (_) {}
            
            if (videoFiles.length === 0) {
                showError('Видео файлы не найдены в торренте');
                return;
            }
            
            // Показываем список файлов
            renderFileList(videoFiles);
            
            // Автоматически выбираем первый (наибольший) файл
            if (videoFiles.length > 0) {
                playFile(videoFiles[0], 0);
            }
            
            elements.loading.style.display = 'none';
            elements.controls.style.display = 'block';
        }
        
        function fetchMediaMetadata(torrentName) {
            // Извлекаем название для поиска из имени торрента
            const searchQuery = extractTitleFromTorrentName(torrentName);
            
            fetch('/api/v1/webtorrent/metadata?query=' + encodeURIComponent(searchQuery))
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
            
            // Готовим видео элемент
            elements.videoPlayer.setAttribute('playsinline', 'true');
            elements.videoPlayer.preload = 'auto';
            
            // Пытаемся воспроизвести через MediaSource
            file.renderTo(elements.videoPlayer, { autoplay: true }, (err) => {
                if (err) {
                    console.warn('renderTo failed, trying fallbacks:', err);
                    // Фолбэк 1: appendTo контейнер (WebTorrent сам создаст video)
                    try {
                        file.appendTo('.player-container', (appendErr, elem) => {
                            if (appendErr) {
                                console.warn('appendTo failed, trying blob URL:', appendErr);
                                // Фолбэк 2: Blob URL
                                try {
                                    file.getBlobURL((blobErr, url) => {
                                        if (blobErr) {
                                            showError('Ошибка воспроизведения (blob): ' + blobErr.message);
                                            return;
                                        }
                                        elements.videoPlayer.src = url;
                                        elements.videoPlayer.style.display = 'block';
                                        const playPromise = elements.videoPlayer.play();
                                        if (playPromise && typeof playPromise.catch === 'function') {
                                            playPromise.catch(() => {});
                                        }
                                    });
                                } catch (e) {
                                    showError('Ошибка воспроизведения: ' + e.message);
                                }
                                return;
                            }
                            // Если appendTo сработал, скрываем наш video, т.к. создан новый
                            elements.videoPlayer.style.display = 'none';
                        });
                    } catch (e) {
                        showError('Ошибка воспроизведения: ' + e.message);
                    }
                } else {
                    elements.videoPlayer.style.display = 'block';
                    const playPromise = elements.videoPlayer.play();
                    if (playPromise && typeof playPromise.catch === 'function') {
                        playPromise.catch(() => {});
                    }
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
            console.error('WebTorrent Error:', message);
        }
        
        // Обработка прогресса загрузки и диагностика пиров
        client.on('torrent', (torrent) => {
            const updateStats = () => {
                const progress = Math.round(torrent.progress * 100);
                const downloadSpeed = (torrent.downloadSpeed / 1024 / 1024).toFixed(1);
                const numPeers = torrent.numPeers || 0;
                elements.loadingProgress.textContent =
                    'Прогресс: ' + progress + '% | Пиры: ' + numPeers + ' | Скорость: ' + downloadSpeed + ' MB/s';
                // Доп. лог для диагностики
                if (progress % 5 === 0) {
                    console.log(`[WebTorrent] ${progress}% peers=${numPeers} speed=${downloadSpeed}MB/s`);
                }
            };
            torrent.on('download', updateStats);
            torrent.on('wire', updateStats);
            torrent.on('noPeers', (announceType) => {
                console.warn('Нет пиров (' + announceType + '). Возможно, требуется VPN или другие WSS-трекеры.');
            });
        });

        // Корректное завершение клиента при уходе со страницы
        window.addEventListener('beforeunload', () => {
            try { client.destroy(() => {}); } catch (_) {}
        });

        // Дополнительные предупреждения
        client.on('warning', (w) => {
            console.warn('WebTorrent warning:', w && w.message ? w.message : w);
        });
        
        // Глобальная обработка ошибок
        client.on('error', (err) => {
            console.error('WebTorrent client error:', err);
            showError('Ошибка торрент клиента: ' + err.message);
        });
        
        // Обработка ошибок торрента
        client.on('torrent', (torrent) => {
            torrent.on('error', (err) => {
                console.error('Torrent error:', err);
                showError('Ошибка торрента: ' + err.message);
            });
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

	// Подготавливаем JSON-строку для безопасной вставки в JS
	mlJSON, _ := json.Marshal(cleanedMagnet)
	data := struct {
		MagnetLinkJSON template.JS
	}{
		MagnetLinkJSON: template.JS(string(mlJSON)),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

// Проверяем есть ли нужные методы в TMDB сервисе
func (h *WebTorrentHandler) checkMethods() {
	// Эти методы должны существовать в TMDBService:
	// - SearchMovies
	// - SearchTV  
	// - GetMovie
	// - GetTVShow
	// - GetTVSeason
}

// isValidMagnetLink проверяет, что ссылка является корректной magnet ссылкой
func isValidMagnetLink(magnetLink string) bool {
	// Проверяем, что ссылка начинается с magnet:
	if len(magnetLink) < 8 || magnetLink[:8] != "magnet:?" {
		return false
	}
	
	// Проверяем наличие обязательных параметров
	if !contains(magnetLink, "xt=urn:btih:") {
		return false
	}
	
	return true
}

// cleanMagnetLink очищает magnet ссылку от лишних символов и нормализует её
func cleanMagnetLink(magnetLink string) string {
	// Убираем лишние пробелы
	cleaned := strings.TrimSpace(magnetLink)
	
	// Убираем переносы строк
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")
	
	// Убираем лишние пробелы внутри ссылки
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	
	// Проверяем, что ссылка начинается с magnet:
	if !strings.HasPrefix(cleaned, "magnet:?") {
		return magnetLink // Возвращаем оригинал если что-то пошло не так
	}
	
	return cleaned
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}