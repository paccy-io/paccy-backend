// trim and validate a string
package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TrimAndValidateString trims whitespace from a string and validates it
func TrimAndValidateString(s string, validate *validator.Validate) (string, error) {
	s = strings.TrimSpace(s)
	if err := validate.Var(s, "required,min=1,max=255"); err != nil {
		return "", err
	}
	return s, nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string, validate *validator.Validate) error {
	if err := validate.Var(email, "required,email"); err != nil {
		return err
	}
	return nil
}

// validate mongo primitive.ObjectID
func ValidateObjectID(id string, validate *validator.Validate) (primitive.ObjectID, error) {
	if err := validate.Var(id, "required,len=24,hexadecimal"); err != nil {
		return primitive.ObjectID{}, err
	}

	return primitive.ObjectIDFromHex(id)
}

// ValidateUUID validates a UUID string
func ValidateUUID(uuidStr string, validate *validator.Validate) (uuid.UUID, error) {
	if err := validate.Var(uuidStr, "required,uuid"); err != nil {
		return uuid.UUID{}, err
	}

	return uuid.Parse(uuidStr)
}