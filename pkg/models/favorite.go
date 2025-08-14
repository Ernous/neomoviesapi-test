package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Favorite struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID     string             `json:"userId" bson:"userId"`
	MediaID    string             `json:"mediaId" bson:"mediaId"`
	MediaType  string             `json:"mediaType" bson:"mediaType"` // "movie" or "tv"
	Title      string             `json:"title" bson:"title"`
	PosterPath string             `json:"posterPath" bson:"posterPath"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
}

type FavoriteRequest struct {
	MediaID    string `json:"mediaId" validate:"required"`
	MediaType  string `json:"mediaType" validate:"required,oneof=movie tv"`
	Title      string `json:"title" validate:"required"`
	PosterPath string `json:"posterPath"`
}