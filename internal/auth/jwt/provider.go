package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/paccy-io/paccy-backend/internal/user/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Provider interface {
	GenerateToken(userID primitive.ObjectID, username string, roles []models.UserRole, verificationLevel models.VerificationLevel) (string, error)
	VerifyToken(tokenStr string) (*Claims, error)
	GenerateEmailVerificationToken(userID primitive.ObjectID) (string, error)
}

type jwtProvider struct {
	secret     string
	expiration time.Duration
	issuer     string
}

func NewJWTProvider(secret string, expiration time.Duration, issuer string) Provider {
	return &jwtProvider{
		secret:     secret,
		expiration: expiration,
		issuer:     issuer,
	}
}

func (p *jwtProvider) GenerateToken(userID primitive.ObjectID, username string, roles []models.UserRole, verificationLevel models.VerificationLevel) (string, error) {
	now := time.Now()

	claims := &Claims{
		ID:                userID,
		Roles:             roles,
		Username:          username,
		VerificationLevel: verificationLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(p.expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secret))
}

// email verification token
func (p *jwtProvider) GenerateEmailVerificationToken(userID primitive.ObjectID) (string, error) {
	// Generate a JWT token for email verification
	claims := &Claims{
		ID:                userID,
		VerificationLevel: models.EmailVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    p.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.secret))
}

func (p *jwtProvider) VerifyToken(tokenStr string) (*Claims, error) {
	// trim any whitespace from the token string
	tokenStr = strings.TrimSpace(tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// İmzalama yöntemi kontrolü
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println("VerifyToken error: unexpected signing method")
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(p.secret), nil
	})

	if err != nil {
		return nil, jwt.ErrSignatureInvalid
	}
	if !token.Valid {
		return nil, jwt.ErrTokenSignatureInvalid
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, jwt.ErrTokenMalformed
}
