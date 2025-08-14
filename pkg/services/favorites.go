package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"neomovies-api/pkg/models"
)

type FavoritesService struct {
	db   *mongo.Database
	tmdb *TMDBService
}

func NewFavoritesService(db *mongo.Database, tmdb *TMDBService) *FavoritesService {
	return &FavoritesService{
		db:   db,
		tmdb: tmdb,
	}
}

func (s *FavoritesService) AddToFavorites(userID, mediaID, mediaType string) error {
	collection := s.db.Collection("favorites")
	
	// Проверяем, не добавлен ли уже в избранное
	filter := bson.M{
		"userId":    userID,
		"mediaId":   mediaID,
		"mediaType": mediaType,
	}
	
	var existingFavorite models.Favorite
	err := collection.FindOne(context.Background(), filter).Decode(&existingFavorite)
	if err == nil {
		// Уже в избранном
		return nil
	}
	
	var title, posterPath string
	
	// Получаем информацию из TMDB в зависимости от типа медиа
	mediaIDInt, err := strconv.Atoi(mediaID)
	if err != nil {
		return fmt.Errorf("invalid media ID: %s", mediaID)
	}
	
	if mediaType == "movie" {
		movie, err := s.tmdb.GetMovie(mediaIDInt, "en-US")
		if err != nil {
			return err
		}
		title = movie.Title
		posterPath = movie.PosterPath
	} else if mediaType == "tv" {
		tv, err := s.tmdb.GetTVShow(mediaIDInt, "en-US")
		if err != nil {
			return err
		}
		title = tv.Name
		posterPath = tv.PosterPath
	} else {
		return fmt.Errorf("invalid media type: %s", mediaType)
	}
	
	// Формируем полный URL для постера
	if posterPath != "" {
		posterPath = fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", posterPath)
	}
	
	favorite := models.Favorite{
		UserID:     userID,
		MediaID:    mediaID,
		MediaType:  mediaType,
		Title:      title,
		PosterPath: posterPath,
		CreatedAt:  time.Now(),
	}
	
	_, err = collection.InsertOne(context.Background(), favorite)
	return err
}

func (s *FavoritesService) RemoveFromFavorites(userID, mediaID, mediaType string) error {
	collection := s.db.Collection("favorites")
	
	filter := bson.M{
		"userId":    userID,
		"mediaId":   mediaID,
		"mediaType": mediaType,
	}
	
	_, err := collection.DeleteOne(context.Background(), filter)
	return err
}

func (s *FavoritesService) GetFavorites(userID string) ([]models.Favorite, error) {
	collection := s.db.Collection("favorites")
	
	filter := bson.M{
		"userId": userID,
	}
	
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	var favorites []models.Favorite
	err = cursor.All(context.Background(), &favorites)
	if err != nil {
		return nil, err
	}
	
	// Возвращаем пустой массив вместо nil если нет избранных
	if favorites == nil {
		favorites = []models.Favorite{}
	}
	
	return favorites, nil
}

func (s *FavoritesService) IsFavorite(userID, mediaID, mediaType string) (bool, error) {
	collection := s.db.Collection("favorites")
	
	filter := bson.M{
		"userId":    userID,
		"mediaId":   mediaID,
		"mediaType": mediaType,
	}
	
	var favorite models.Favorite
	err := collection.FindOne(context.Background(), filter).Decode(&favorite)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}