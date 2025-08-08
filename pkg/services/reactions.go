package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"neomovies-api/pkg/config"
	"neomovies-api/pkg/models"
)

type ReactionsService struct {
	db     *mongo.Database
	client *http.Client
}

func NewReactionsService(db *mongo.Database) *ReactionsService {
	return &ReactionsService{
		db:     db,
		client: &http.Client{},
	}
}

var validReactions = []string{"fire", "nice", "think", "bore", "shit"}

// Получить счетчики реакций для медиа из внешнего API (cub.rip)
func (s *ReactionsService) GetReactionCounts(mediaType, mediaID string) (*models.ReactionCounts, error) {
	cubID := fmt.Sprintf("%s_%s", mediaType, mediaID)
	
	resp, err := s.client.Get(fmt.Sprintf("%s/reactions/get/%s", config.CubAPIBaseURL, cubID))
	if err != nil {
		return &models.ReactionCounts{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &models.ReactionCounts{}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &models.ReactionCounts{}, nil
	}

	var response struct {
		Result []struct {
			Type    string `json:"type"`
			Counter int    `json:"counter"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return &models.ReactionCounts{}, nil
	}

	counts := &models.ReactionCounts{}
	for _, reaction := range response.Result {
		switch reaction.Type {
		case "fire":
			counts.Fire = reaction.Counter
		case "nice":
			counts.Nice = reaction.Counter
		case "think":
			counts.Think = reaction.Counter
		case "bore":
			counts.Bore = reaction.Counter
		case "shit":
			counts.Shit = reaction.Counter
		}
	}

	return counts, nil
}

func (s *ReactionsService) GetMyReaction(userID, mediaType, mediaID string) (string, error) {
	collection := s.db.Collection("reactions")
	ctx := context.Background()

	var result struct{ Type string `bson:"type"` }
	err := collection.FindOne(ctx, bson.M{
		"userId":    userID,
		"mediaType": mediaType,
		"mediaId":   mediaID,
	}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	return result.Type, nil
}

func (s *ReactionsService) SetReaction(userID, mediaType, mediaID, reactionType string) error {
	if !s.isValidReactionType(reactionType) {
		return fmt.Errorf("invalid reaction type")
	}

	collection := s.db.Collection("reactions")
	ctx := context.Background()

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"userId": userID, "mediaType": mediaType, "mediaId": mediaID},
		bson.M{"$set": bson.M{"type": reactionType, "updatedAt": time.Now()}},
		options.Update().SetUpsert(true),
	)
	if err == nil {
		go s.sendReactionToCub(fmt.Sprintf("%s_%s", mediaType, mediaID), reactionType)
	}
	return err
}

func (s *ReactionsService) RemoveReaction(userID, mediaType, mediaID string) error {
	collection := s.db.Collection("reactions")
	ctx := context.Background()

	_, err := collection.DeleteOne(ctx, bson.M{
		"userId":    userID,
		"mediaType": mediaType,
		"mediaId":   mediaID,
	})

	fullMediaID := fmt.Sprintf("%s_%s", mediaType, mediaID)
	go s.sendReactionToCub(fullMediaID, "remove")

	return err
}

// Получить все реакции пользователя
func (s *ReactionsService) GetUserReactions(userID string, limit int) ([]models.Reaction, error) {
	collection := s.db.Collection("reactions")

	ctx := context.Background()
	cursor, err := collection.Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reactions []models.Reaction
	if err := cursor.All(ctx, &reactions); err != nil {
		return nil, err
	}

	return reactions, nil
}

func (s *ReactionsService) isValidReactionType(reactionType string) bool {
	for _, valid := range validReactions {
		if valid == reactionType {
			return true
		}
	}
	return false
}

// Отправка реакции в cub.rip API (асинхронно)
func (s *ReactionsService) sendReactionToCub(mediaID, reactionType string) {
	url := fmt.Sprintf("%s/reactions/set", config.CubAPIBaseURL)
	
	data := map[string]string{
		"mediaId": mediaID,
		"type":    reactionType,
	}

	_, err := json.Marshal(data)
	if err != nil {
		return
	}

	resp, err := s.client.Get(fmt.Sprintf("%s?mediaId=%s&type=%s", url, mediaID, reactionType))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Reaction sent to cub.rip: %s - %s\n", mediaID, reactionType)
	}
}