# Neo Movies API

REST API для поиска и получения информации о фильмах, использующий TMDB API.

## Особенности

- Поиск фильмов
- Информация о фильмах
- Популярные фильмы
- Топ рейтинговые фильмы
- Предстоящие фильмы
- Swagger документация
- Поддержка русского языка

## 🛠 Быстрый старт

### Локальная разработка

1. **Клонирование репозитория**
```bash
git clone <your-repo>
cd neomovies-api
```

2. **Создание .env файла**
```bash
cp .env.example .env
# Заполните необходимые переменные
```

3. **Установка зависимостей**
```bash
go mod download
```

4. **Запуск**
```bash
go run main.go
```

API будет доступен на `http://localhost:3000`

### Деплой на Vercel

1. **Подключите репозиторий к Vercel**
2. **Настройте переменные окружения** (см. список ниже)
3. **Деплой произойдет автоматически**

## ⚙️ Переменные окружения

```bash
# Обязательные
MONGO_URI=
MONGO_DB_NAME=database
TMDB_ACCESS_TOKEN=
JWT_SECRET=

# Сервис
PORT=3000
BASE_URL=http://localhost:3000
NODE_ENV=development

# Email (Gmail)
GMAIL_USER=
GMAIL_APP_PASSWORD=

# Плееры
LUMEX_URL=
ALLOHA_TOKEN=

# Торренты (RedAPI)
REDAPI_BASE_URL=http://redapi.cfhttp.top
REDAPI_KEY=

# Google OAuth
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:3000/api/v1/auth/google/callback
```

## 📋 API Endpoints

### 🔓 Публичные маршруты

```http
# Система
GET  /api/v1/health                          # Проверка состояния

# Аутентификация
POST /api/v1/auth/register                   # Регистрация (отправка кода)
POST /api/v1/auth/verify                     # Подтверждение email кодом
POST /api/v1/auth/resend-code               # Повторная отправка кода
POST /api/v1/auth/login                      # Авторизация
GET  /api/v1/auth/google/login               # Начало авторизации через Google (redirect)
GET  /api/v1/auth/google/callback            # Коллбек Google OAuth (возвращает JWT)

# Поиск и категории
GET  /search/multi                           # Мультипоиск
GET  /api/v1/categories                      # Список категорий
GET  /api/v1/categories/{id}/movies          # Фильмы по категории

# Фильмы
GET  /api/v1/movies/search                   # Поиск фильмов
GET  /api/v1/movies/popular                  # Популярные
GET  /api/v1/movies/top-rated                # Топ-рейтинговые
GET  /api/v1/movies/upcoming                 # Предстоящие
GET  /api/v1/movies/now-playing              # В прокате
GET  /api/v1/movies/{id}                     # Детали фильма
GET  /api/v1/movies/{id}/recommendations     # Рекомендации
GET  /api/v1/movies/{id}/similar             # Похожие

# Сериалы
GET  /api/v1/tv/search                       # Поиск сериалов
GET  /api/v1/tv/popular                      # Популярные
GET  /api/v1/tv/top-rated                    # Топ-рейтинговые
GET  /api/v1/tv/on-the-air                   # В эфире
GET  /api/v1/tv/airing-today                 # Сегодня в эфире
GET  /api/v1/tv/{id}                         # Детали сериала
GET  /api/v1/tv/{id}/recommendations         # Рекомендации
GET  /api/v1/tv/{id}/similar                 # Похожие

# Плееры
GET  /api/v1/players/alloha/{imdb_id}          # Alloha плеер по IMDb ID
GET  /api/v1/players/lumex/{imdb_id}           # Lumex плеер по IMDb ID

# Торренты
GET  /api/v1/torrents/search/{imdbId}        # Поиск торрентов

# Реакции (публичные)
GET  /api/v1/reactions/{mediaType}/{mediaId}/counts    # Счетчики реакций

# Изображения
GET  /api/v1/images/{size}/{path}            # Прокси TMDB изображений
```

### 🔒 Приватные маршруты (требуют JWT)

```http
# Профиль
GET  /api/v1/auth/profile                    # Профиль пользователя
PUT  /api/v1/auth/profile                    # Обновление профиля

# Избранное
GET  /api/v1/favorites                       # Список избранного
POST /api/v1/favorites/{id}                  # Добавить в избранное
DELETE /api/v1/favorites/{id}                # Удалить из избранного

# Реакции (приватные)
GET  /api/v1/reactions/{mediaType}/{mediaId}/my-reaction # Моя реакция
POST /api/v1/reactions/{mediaType}/{mediaId}           # Установить реакцию
DELETE /api/v1/reactions/{mediaType}/{mediaId}         # Удалить реакцию
GET  /api/v1/reactions/my                              # Все мои реакции
```

## 📖 Примеры использования

### Регистрация и верификация

```bash
# 1. Регистрация
curl -X POST https://api.neomovies.ru/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'

# Ответ: {"success": true, "message": "Registered. Check email for verification code."}

# 2. Подтверждение email (код из письма)
curl -X POST https://api.neomovies.ru/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "code": "123456"
  }'

# 3. Авторизация
curl -X POST https://api.neomovies.ru/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Поиск фильмов

```bash
# Поиск фильмов
curl "https://api.neomovies.ru/api/v1/movies/search?query=marvel&page=1"

# Детали фильма
curl "https://api.neomovies.ru/api/v1/movies/550"

# Добавить в избранное (с JWT токеном)
curl -X POST https://api.neomovies.ru/api/v1/favorites/550 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Поиск торрентов

```bash
# Поиск торрентов для фильма "Побег из Шоушенка"
curl "https://api.neomovies.ru/api/v1/torrents/search/tt0111161?type=movie&quality=1080p"
```

## 🎨 Документация API

Интерактивная документация доступна по адресу:

**🔗 https://api.neomovies.ru/**

## ☁️ Деплой на Vercel

1. **Подключите репозиторий к Vercel**
2. **Настройте Environment Variables в Vercel Dashboard:**
3. **Деплой автоматически запустится!**

## 🏗 Архитектура

```
├── main.go                 # Точка входа приложения
├── api/
│   └── index.go           # Vercel serverless handler
├── pkg/                   # Публичные пакеты (совместимо с Vercel)
│   ├── config/           # Конфигурация с поддержкой альтернативных env vars
│   ├── database/         # Подключение к MongoDB
│   ├── middleware/       # JWT, CORS, логирование
│   ├── models/          # Структуры данных
│   ├── services/        # Бизнес-логика
│   └── handlers/        # HTTP обработчики
├── vercel.json          # Конфигурация Vercel
└── go.mod              # Go модули
```

## 🔧 Технологии

- **Go 1.21** - основной язык
- **Gorilla Mux** - HTTP роутер
- **MongoDB** - база данных
- **JWT** - аутентификация
- **TMDB API** - данные о фильмах
- **Gmail SMTP** - email уведомления
- **Vercel** - деплой и хостинг

## 🚀 Производительность

По сравнению с Node.js версией:
- **3x быстрее** обработка запросов
- **50% меньше** потребление памяти
- **Конкурентность** благодаря горутинам
- **Типобезопасность** предотвращает ошибки

## 🤝 Contribution

1. Форкните репозиторий
2. Создайте feature-ветку (`git checkout -b feature/amazing-feature`)
3. Коммитьте изменения (`git commit -m 'Add amazing feature'`)
4. Пушните в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

Apache License 2.0 - подробности в файле [LICENSE](LICENSE)

---

Made with <3 by Foxix