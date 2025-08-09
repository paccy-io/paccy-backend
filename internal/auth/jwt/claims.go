package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Claims struct {
	ID                primitive.ObjectID       `json:"user_id"` // it's ObjectID not UUID
	Username          string                   `json:"username"`
	Roles             []models.UserRole        `json:"roles"`
	VerificationLevel models.VerificationLevel `json:"verification_level"`
	jwt.RegisteredClaims
}
