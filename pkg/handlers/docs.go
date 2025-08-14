package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeDocs(w, r)
}

func (h *DocsHandler) RedirectToDocs(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
}

func (h *DocsHandler) GetOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	baseURL := determineBaseURL(r)

	spec := getOpenAPISpecWithURL(baseURL)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Origin, X-Requested-With")
	json.NewEncoder(w).Encode(spec)
}

func (h *DocsHandler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	baseURL := determineBaseURL(r)

	htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
		SpecURL: fmt.Sprintf("%s/openapi.json", baseURL),
		CustomOptions: scalar.CustomOptions{
			PageTitle: "Neo Movies API Documentation",
		},
		DarkMode: true,
	})

	if err != nil {
		fmt.Printf("Error generating documentation: %v", err)
		http.Error(w, fmt.Sprintf("Error generating documentation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprintln(w, htmlContent)
}

func determineBaseURL(r *http.Request) string {
	if envBase := os.Getenv("BASE_URL"); envBase != "" {
		return envBase
	}

	// Defaults
	proto := ""
	host := r.Host

	// RFC 7239 Forwarded header: e.g. "for=1.2.3.4; proto=https; host=example.com"
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		for _, part := range strings.Split(fwd, ";") {
			kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
			switch key {
			case "proto":
				if proto == "" {
					proto = strings.ToLower(val)
				}
			case "host":
				if val != "" {
					host = val
				}
			}
		}
	}

	// Fallback to X-Forwarded-* headers
	if proto == "" {
		if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
			proto = strings.ToLower(strings.TrimSpace(strings.Split(p, ",")[0]))
		}
	}
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		host = strings.TrimSpace(strings.Split(xfh, ",")[0])
	}

	// Last resort: infer from TLS
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	// Respect X-Forwarded-Port if host has no explicit port and it is non-default
	if xfp := r.Header.Get("X-Forwarded-Port"); xfp != "" && !strings.Contains(host, ":") {
		isDefault := (proto == "http" && xfp == "80") || (proto == "https" && xfp == "443")
		if !isDefault {
			host = host + ":" + xfp
		}
	}

	return fmt.Sprintf("%s://%s", proto, host)
}

type OpenAPISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    Info                   `json:"info"`
	Servers []Server               `json:"servers"`
	Paths   map[string]interface{} `json:"paths"`
	Components Components           `json:"components"`
}

type Info struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
}

type Contact struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
	Schemas         map[string]interface{}    `json:"schemas"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

func getOpenAPISpecWithURL(baseURL string) *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       "Neo Movies API",
			Description: "Современный API для поиска фильмов и сериалов с интеграцией TMDB и поддержкой авторизации",
			Version:     "2.0.0",
			Contact: Contact{
				Name: "API Support",
				URL:  "https://github.com/your-username/neomovies-api-go",
			},
		},
		Servers: []Server{
			{
				URL:         baseURL,
				Description: "Production server",
			},
		},
		Paths: map[string]interface{}{
			"/api/v1/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health Check",
					"description": "Проверка работоспособности API",
					"tags": []string{"Health"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "API работает корректно",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/APIResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/search/multi": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Мультипоиск",
					"description": "Поиск фильмов, сериалов и актеров",
					"tags": []string{"Search"},
					"parameters": []map[string]interface{}{
						{
							"name": "query",
							"in": "query",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Поисковый запрос",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
							"description": "Номер страницы",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Результаты поиска",
						},
					},
				},
			},
			"/api/v1/categories": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Получить категории",
					"description": "Получение списка категорий фильмов",
					"tags": []string{"Categories"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список категорий",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{"$ref": "#/components/schemas/Category"},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/categories/{id}/movies": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Фильмы по категории",
					"description": "Получение фильмов по категории",
					"tags": []string{"Categories"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID категории",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Фильмы категории",
						},
					},
				},
			},
			"/api/v1/categories/{id}/media": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Медиа по категории",
					"description": "Получение фильмов или сериалов по категории",
					"tags": []string{"Categories"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID категории",
						},
						{
							"name": "type",
							"in": "query",
							"required": false,
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"movie", "tv"},
								"default": "movie",
							},
							"description": "Тип медиа: movie или tv",
						},
						{
							"name": "page",
							"in": "query",
							"required": false,
							"schema": map[string]interface{}{
								"type": "integer",
								"default": 1,
							},
							"description": "Номер страницы",
						},
						{
							"name": "language",
							"in": "query",
							"required": false,
							"schema": map[string]interface{}{
								"type": "string",
								"default": "ru-RU",
							},
							"description": "Язык ответа",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Медиа по категории",
						},
					},
				},
			},
			"/api/v1/players/alloha/{imdb_id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Плеер Alloha",
					"description": "Получение плеера Alloha по IMDb ID",
					"tags": []string{"Players"},
					"parameters": []map[string]interface{}{
						{
							"name": "imdb_id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "IMDb ID фильма",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Данные плеера",
						},
					},
				},
			},
			"/api/v1/players/lumex/{imdb_id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Плеер Lumex",
					"description": "Получение плеера Lumex по IMDb ID",
					"tags": []string{"Players"},
					"parameters": []map[string]interface{}{
						{
							"name": "imdb_id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "IMDb ID фильма",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Данные плеера",
						},
					},
				},
			},
			"/api/v1/players/vibix/{imdb_id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Vibix плеер по IMDb ID",
					"description": "Возвращает HTML-страницу с iframe Vibix для указанного IMDb ID",
					"tags":        []string{"Players"},
					"parameters": []map[string]interface{}{
						{
							"name":        "imdb_id",
							"in":          "path",
							"required":    true,
							"schema":      map[string]string{"type": "string"},
							"description": "IMDb ID, например tt0133093",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "HTML со встроенным Vibix плеером",
							"content": map[string]interface{}{
								"text/html": map[string]interface{}{},
							},
						},
						"404": map[string]interface{}{"description": "Фильм не найден"},
						"503": map[string]interface{}{"description": "VIBIX_TOKEN не настроен"},
					},
				},
			},
			"/api/v1/webtorrent/player": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "WebTorrent плеер",
					"description": "Открытие WebTorrent плеера с магнет ссылкой. Плеер работает полностью на стороне клиента.",
					"tags": []string{"WebTorrent"},
					"parameters": []map[string]interface{}{
						{
							"name": "magnet",
							"in": "query",
							"required": false,
							"schema": map[string]string{"type": "string"},
							"description": "Магнет ссылка торрента",
						},
						{
							"name": "X-Magnet-Link",
							"in": "header",
							"required": false,
							"schema": map[string]string{"type": "string"},
							"description": "Магнет ссылка через заголовок (альтернативный способ)",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "HTML страница с WebTorrent плеером",
							"content": map[string]interface{}{
								"text/html": map[string]interface{}{
									"schema": map[string]string{"type": "string"},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Отсутствует магнет ссылка",
						},
					},
				},
			},
			"/api/v1/webtorrent/metadata": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Метаданные медиа",
					"description": "Получение метаданных фильма или сериала по названию для WebTorrent плеера",
					"tags": []string{"WebTorrent"},
					"parameters": []map[string]interface{}{
						{
							"name": "query",
							"in": "query",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Название для поиска (извлеченное из торрента)",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Метаданные найдены",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/WebTorrentMetadata",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Отсутствует параметр query",
						},
						"404": map[string]interface{}{
							"description": "Метаданные не найдены",
						},
					},
				},
			},
			"/api/v1/torrents/search/{imdbId}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Поиск торрентов",
					"description": "Поиск торрентов по IMDB ID",
					"tags":        []string{"Torrents"},
					"parameters": []map[string]interface{}{
						{
							"name":        "imdbId",
							"in":          "path",
							"required":    true,
							"schema":      map[string]string{"type": "string"},
							"description": "IMDB ID фильма или сериала",
						},
						{
							"name":        "type",
							"in":          "query",
							"required":    true,
							"schema":      map[string]interface{}{"type": "string", "enum": []string{"movie", "tv", "serial"}},
							"description": "Тип контента: movie (фильм) или tv/serial (сериал)",
						},
						{
							"name":        "season",
							"in":          "query",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer"},
							"description": "Номер сезона (для сериалов)",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Результаты поиска торрентов",
						},
					},
				},
			},
			"/api/v1/reactions/{mediaType}/{mediaId}/counts": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Количество реакций",
					"description": "Получение количества реакций для медиа",
					"tags": []string{"Reactions"},
					"parameters": []map[string]interface{}{
						{
							"name": "mediaType",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Тип медиа (movie/tv)",
						},
						{
							"name": "mediaId",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "ID медиа",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Количество реакций",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ReactionCounts",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/images/{size}/{path}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Изображения",
					"description": "Прокси для изображений TMDB",
					"tags": []string{"Images"},
					"parameters": []map[string]interface{}{
						{
							"name": "size",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Размер изображения",
						},
						{
							"name": "path",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Путь к изображению",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Изображение",
							"content": map[string]interface{}{
								"image/*": map[string]interface{}{},
							},
						},
					},
				},
			},
			"/api/v1/auth/register": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Регистрация пользователя",
					"description": "Создание нового аккаунта пользователя",
					"tags": []string{"Authentication"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/RegisterRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Пользователь успешно зарегистрирован",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/AuthResponse",
									},
								},
							},
						},
						"409": map[string]interface{}{
							"description": "Пользователь с таким email уже существует",
						},
					},
				},
			},
			"/api/v1/auth/verify": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Подтверждение email",
					"description": "Подтверждение email пользователя с помощью кода",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"required": []string{"email", "code"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":        "string",
											"format":      "email",
											"description": "Email пользователя",
											"example":     "user@example.com",
										},
										"code": map[string]interface{}{
											"type":        "string",
											"description": "6-значный код верификации",
											"example":     "123456",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Email успешно подтвержден",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{
												"type": "boolean",
											},
											"message": map[string]interface{}{
												"type": "string",
											},
										},
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Неверный или истекший код",
						},
					},
				},
			},
			"/api/v1/auth/resend-code": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Повторная отправка кода",
					"description": "Повторная отправка кода верификации на email",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"required": []string{"email"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":        "string",
											"format":      "email",
											"description": "Email пользователя",
											"example":     "user@example.com",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Код отправлен на email",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{
												"type": "boolean",
											},
											"message": map[string]interface{}{
												"type": "string",
											},
										},
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Email уже подтвержден или пользователь не найден",
						},
					},
				},
			},
			"/api/v1/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Авторизация пользователя",
					"description": "Получение JWT токена для доступа к приватным эндпоинтам",
					"tags": []string{"Authentication"},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/LoginRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Успешная авторизация",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/AuthResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Неверный email или пароль",
						},
					},
				},
			},
			"/api/v1/auth/profile": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Получить профиль пользователя",
					"description": "Получение информации о текущем пользователе",
					"tags": []string{"Authentication"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Информация о пользователе",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
					},
				},
				"put": map[string]interface{}{
					"summary": "Обновить профиль пользователя",
					"description": "Обновление информации о пользователе",
					"tags": []string{"Authentication"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Профиль успешно обновлен",
						},
					},
				},
				"delete": map[string]interface{}{
					"summary": "Удалить аккаунт пользователя",
					"description": "Полное и безвозвратное удаление аккаунта пользователя и всех связанных с ним данных (избранное, реакции)",
					"tags": []string{"Authentication"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Аккаунт успешно удален",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"success": map[string]interface{}{"type": "boolean"},
											"message": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Неавторизованный запрос",
						},
						"500": map[string]interface{}{
							"description": "Внутренняя ошибка сервера",
						},
					},
				},
			},
			"/api/v1/movies/search": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Поиск фильмов",
					"description": "Поиск фильмов по названию с поддержкой фильтров",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "query",
							"in": "query",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Поисковый запрос",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
							"description": "Номер страницы",
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
							"description": "Язык ответа",
						},
						{
							"name": "year",
							"in": "query",
							"schema": map[string]string{"type": "integer"},
							"description": "Год выпуска",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Результаты поиска фильмов",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/MovieSearchResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/movies/popular": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Популярные фильмы",
					"description": "Получение списка популярных фильмов",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список популярных фильмов",
						},
					},
				},
			},
			"/api/v1/movies/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Получить фильм по ID",
					"description": "Подробная информация о фильме",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID фильма в TMDB",
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Информация о фильме",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/Movie",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/favorites": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Получить избранное",
					"description": "Список избранных фильмов и сериалов пользователя",
					"tags": []string{"Favorites"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список избранного",
						},
					},
				},
			},
			"/api/v1/favorites/{id}": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Добавить в избранное",
					"description": "Добавление фильма или сериала в избранное",
					"tags": []string{"Favorites"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "ID медиа",
						},
						{
							"name": "type",
							"in": "query",
							"required": false,
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"movie", "tv"},
								"default": "movie",
							},
							"description": "Тип медиа: movie или tv",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Добавлено в избранное",
						},
					},
				},
				"delete": map[string]interface{}{
					"summary": "Удалить из избранного",
					"description": "Удаление фильма или сериала из избранного",
					"tags": []string{"Favorites"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "ID медиа",
						},
						{
							"name": "type",
							"in": "query",
							"required": false,
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"movie", "tv"},
								"default": "movie",
							},
							"description": "Тип медиа: movie или tv",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Удалено из избранного",
						},
					},
				},
			},
			"/api/v1/favorites/{id}/check": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Проверить избранное",
					"description": "Проверка, находится ли медиа в избранном",
					"tags": []string{"Favorites"},
					"security": []map[string][]string{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "ID медиа",
						},
						{
							"name": "type",
							"in": "query",
							"required": false,
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"movie", "tv"},
								"default": "movie",
							},
							"description": "Тип медиа: movie или tv",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Статус избранного",
						},
					},
				},
			},
			"/api/v1/movies/top-rated": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Топ рейтинг фильмов",
					"description": "Получение списка фильмов с высоким рейтингом",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список фильмов с высоким рейтингом",
						},
					},
				},
			},
			"/api/v1/movies/upcoming": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Скоро в прокате",
					"description": "Получение списка фильмов, которые скоро выйдут в прокат",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список фильмов, которые скоро выйдут",
						},
					},
				},
			},
			"/api/v1/movies/now-playing": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Сейчас в прокате",
					"description": "Получение списка фильмов, которые сейчас в прокате",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список фильмов в прокате",
						},
					},
				},
			},
			"/api/v1/movies/{id}/recommendations": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Рекомендации фильмов",
					"description": "Получение рекомендаций фильмов на основе выбранного",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID фильма в TMDB",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Рекомендуемые фильмы",
						},
					},
				},
			},
			"/api/v1/movies/{id}/similar": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Похожие фильмы",
					"description": "Получение похожих фильмов",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID фильма в TMDB",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Похожие фильмы",
						},
					},
				},
			},
			"/api/v1/tv/search": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Поиск сериалов",
					"description": "Поиск сериалов по названию",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "query",
							"in": "query",
							"required": true,
							"schema": map[string]string{"type": "string"},
							"description": "Поисковый запрос",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
							"description": "Номер страницы",
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
							"description": "Язык ответа",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Результаты поиска сериалов",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/TVSearchResponse",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/tv/popular": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Популярные сериалы",
					"description": "Получение списка популярных сериалов",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список популярных сериалов",
						},
					},
				},
			},
			"/api/v1/tv/top-rated": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Топ рейтинг сериалов",
					"description": "Получение списка сериалов с высоким рейтингом",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список сериалов с высоким рейтингом",
						},
					},
				},
			},
			"/api/v1/tv/on-the-air": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "В эфире",
					"description": "Получение списка сериалов, которые сейчас в эфире",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список сериалов в эфире",
						},
					},
				},
			},
			"/api/v1/tv/airing-today": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Сегодня в эфире",
					"description": "Получение списка сериалов, которые выходят сегодня",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Список сериалов, выходящих сегодня",
						},
					},
				},
			},
			"/api/v1/tv/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Получить сериал по ID",
					"description": "Подробная информация о сериале",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID сериала в TMDB",
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Информация о сериале",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/TVSeries",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/tv/{id}/recommendations": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Рекомендации сериалов",
					"description": "Получение рекомендаций сериалов на основе выбранного",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID сериала в TMDB",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Рекомендуемые сериалы",
						},
					},
				},
			},
			"/api/v1/tv/{id}/similar": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Похожие сериалы",
					"description": "Получение похожих сериалов",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID сериала в TMDB",
						},
						{
							"name": "page",
							"in": "query",
							"schema": map[string]string{"type": "integer", "default": "1"},
						},
						{
							"name": "language",
							"in": "query",
							"schema": map[string]string{"type": "string", "default": "ru-RU"},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Похожие сериалы",
						},
					},
				},
			},
			"/api/v1/movies/{id}/external-ids": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Внешние идентификаторы фильма",
					"description": "Получить внешние ID (IMDb, TVDB, Facebook и др.) для фильма по TMDB ID",
					"tags": []string{"Movies"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID фильма в TMDB",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Внешние идентификаторы фильма",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ExternalIDs",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/tv/{id}/external-ids": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Внешние идентификаторы сериала",
					"description": "Получить внешние ID (IMDb, TVDB, Facebook и др.) для сериала по TMDB ID",
					"tags": []string{"TV Series"},
					"parameters": []map[string]interface{}{
						{
							"name": "id",
							"in": "path",
							"required": true,
							"schema": map[string]string{"type": "integer"},
							"description": "ID сериала в TMDB",
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Внешние идентификаторы сериала",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ExternalIDs",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/auth/google/login": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Google OAuth: начало",
					"description": "Редирект на страницу авторизации Google",
					"tags": []string{"Authentication"},
					"responses": map[string]interface{}{
						"302": map[string]interface{}{"description": "Redirect to Google"},
						"400": map[string]interface{}{"description": "OAuth не сконфигурирован"},
					},
				},
			},
			"/api/v1/auth/google/callback": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Google OAuth: коллбек",
					"description": "Обработка кода авторизации и выдача JWT",
					"tags": []string{"Authentication"},
					"parameters": []map[string]interface{}{
						{"name": "state", "in": "query", "required": true, "schema": map[string]string{"type": "string"}},
						{"name": "code", "in": "query", "required": true, "schema": map[string]string{"type": "string"}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Успешная авторизация через Google",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{"$ref": "#/components/schemas/AuthResponse"},
								},
							},
						},
						"400": map[string]interface{}{"description": "Неверный state или ошибка обмена кода"},
					},
				},
			},
		},
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
			Schemas: map[string]interface{}{
				"APIResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]string{"type": "boolean"},
						"data": map[string]string{"type": "object"},
						"message": map[string]string{"type": "string"},
						"error": map[string]string{"type": "string"},
					},
				},
				"RegisterRequest": map[string]interface{}{
					"type": "object",
					"required": []string{"email", "password", "name"},
					"properties": map[string]interface{}{
						"email": map[string]interface{}{
							"type": "string",
							"format": "email",
							"example": "user@example.com",
						},
						"password": map[string]interface{}{
							"type": "string",
							"minLength": 6,
							"example": "password123",
						},
						"name": map[string]interface{}{
							"type": "string",
							"example": "Иван Иванов",
						},
					},
				},
				"LoginRequest": map[string]interface{}{
					"type": "object",
					"required": []string{"email", "password"},
					"properties": map[string]interface{}{
						"email": map[string]interface{}{
							"type": "string",
							"format": "email",
							"example": "user@example.com",
						},
						"password": map[string]interface{}{
							"type": "string",
							"example": "password123",
						},
					},
				},
				"AuthResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"token": map[string]string{"type": "string"},
						"user": map[string]interface{}{"$ref": "#/components/schemas/User"},
					},
				},
				"User": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "string"},
						"email": map[string]string{"type": "string"},
						"name": map[string]string{"type": "string"},
						"avatar": map[string]string{"type": "string"},
						"favorites": map[string]interface{}{
							"type": "array",
							"items": map[string]string{"type": "string"},
						},
						"created_at": map[string]interface{}{
							"type": "string",
							"format": "date-time",
						},
						"updated_at": map[string]interface{}{
							"type": "string",
							"format": "date-time",
						},
					},
				},
				"Movie": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "integer"},
						"title": map[string]string{"type": "string"},
						"original_title": map[string]string{"type": "string"},
						"overview": map[string]string{"type": "string"},
						"poster_path": map[string]string{"type": "string"},
						"backdrop_path": map[string]string{"type": "string"},
						"release_date": map[string]string{"type": "string"},
						"vote_average": map[string]string{"type": "number"},
						"vote_count": map[string]string{"type": "integer"},
						"popularity": map[string]string{"type": "number"},
						"adult": map[string]string{"type": "boolean"},
						"original_language": map[string]string{"type": "string"},
					},
				},
				"MovieSearchResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"page": map[string]string{"type": "integer"},
						"results": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{"$ref": "#/components/schemas/Movie"},
						},
						"total_pages": map[string]string{"type": "integer"},
						"total_results": map[string]string{"type": "integer"},
					},
				},
				"TVSeries": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "integer"},
						"name": map[string]string{"type": "string"},
						"original_name": map[string]string{"type": "string"},
						"overview": map[string]string{"type": "string"},
						"poster_path": map[string]string{"type": "string"},
						"backdrop_path": map[string]string{"type": "string"},
						"first_air_date": map[string]string{"type": "string"},
						"vote_average": map[string]string{"type": "number"},
						"vote_count": map[string]string{"type": "integer"},
						"popularity": map[string]string{"type": "number"},
						"original_language": map[string]string{"type": "string"},
						"number_of_seasons": map[string]string{"type": "integer"},
						"number_of_episodes": map[string]string{"type": "integer"},
						"status": map[string]string{"type": "string"},
					},
				},
				"TVSearchResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"page": map[string]string{"type": "integer"},
						"results": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{"$ref": "#/components/schemas/TVSeries"},
						},
						"total_pages": map[string]string{"type": "integer"},
						"total_results": map[string]string{"type": "integer"},
					},
				},
				"Category": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "integer"},
						"name": map[string]string{"type": "string"},
						"description": map[string]string{"type": "string"},
					},
				},
				"Player": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]string{"type": "string"},
						"title": map[string]string{"type": "string"},
						"quality": map[string]string{"type": "string"},
						"type": map[string]string{"type": "string"},
					},
				},
				"Torrent": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]string{"type": "string"},
						"size": map[string]string{"type": "string"},
						"seeds": map[string]string{"type": "integer"},
						"peers": map[string]string{"type": "integer"},
						"magnet": map[string]string{"type": "string"},
						"hash": map[string]string{"type": "string"},
					},
				},
				"Reaction": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]string{"type": "string"},
						"count": map[string]string{"type": "integer"},
					},
				},
				"ReactionCounts": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"like": map[string]string{"type": "integer"},
						"dislike": map[string]string{"type": "integer"},
						"love": map[string]string{"type": "integer"},
					},
				},
				"ExternalIDs": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "integer"},
						"imdb_id": map[string]string{"type": "string"},
						"tvdb_id": map[string]string{"type": "integer"},
						"wikidata_id": map[string]string{"type": "string"},
						"facebook_id": map[string]string{"type": "string"},
						"instagram_id": map[string]string{"type": "string"},
						"twitter_id": map[string]string{"type": "string"},
					},
				},
				"WebTorrentMetadata": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "integer"},
						"title": map[string]string{"type": "string"},
						"type": map[string]interface{}{
							"type": "string",
							"enum": []string{"movie", "tv"},
						},
						"year": map[string]string{"type": "integer"},
						"posterPath": map[string]string{"type": "string"},
						"backdropPath": map[string]string{"type": "string"},
						"overview": map[string]string{"type": "string"},
						"runtime": map[string]string{"type": "integer"},
						"genres": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/Genre",
							},
						},
						"seasons": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/SeasonMetadata",
							},
						},
						"episodes": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/EpisodeMetadata",
							},
						},
					},
				},
				"SeasonMetadata": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"seasonNumber": map[string]string{"type": "integer"},
						"name": map[string]string{"type": "string"},
						"episodes": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"$ref": "#/components/schemas/EpisodeMetadata",
							},
						},
					},
				},
				"EpisodeMetadata": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"episodeNumber": map[string]string{"type": "integer"},
						"seasonNumber": map[string]string{"type": "integer"},
						"name": map[string]string{"type": "string"},
						"overview": map[string]string{"type": "string"},
						"runtime": map[string]string{"type": "integer"},
						"stillPath": map[string]string{"type": "string"},
					},
				},
				"Genre": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]string{"type": "integer"},
						"name": map[string]string{"type": "string"},
					},
				},
			},
		},
	}
}
