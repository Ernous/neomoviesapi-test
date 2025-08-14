#!/bin/bash

echo "🚀 Запуск NeoMovies API в режиме разработки..."

# Проверяем, установлен ли Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Установите Go для продолжения."
    exit 1
fi

# Проверяем, есть ли файл .env
if [ ! -f .env ]; then
    echo "⚠️  Файл .env не найден. Создаем базовый .env файл..."
    cat > .env << EOF
# NeoMovies API Configuration
PORT=8080
MONGODB_URI=mongodb://localhost:27017
MONGODB_NAME=neomovies
JWT_SECRET=your-secret-key-here
TMDB_ACCESS_TOKEN=your-tmdb-token-here
BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback
RED_API_BASE_URL=https://api.redapi.ru
RED_API_KEY=your-red-api-key
EOF
    echo "✅ Создан базовый .env файл. Отредактируйте его перед запуском."
fi

# Компилируем проект
echo "🔨 Компиляция проекта..."
go build -o neomovies-api .

if [ $? -eq 0 ]; then
    echo "✅ Компиляция успешна!"
    
    # Запускаем сервер
    echo "🌐 Запуск сервера на http://localhost:8080"
    echo "📝 Логи сервера:"
    echo "----------------------------------------"
    ./neomovies-api
else
    echo "❌ Ошибка компиляции!"
    exit 1
fi