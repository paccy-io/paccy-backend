package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseModel struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UUID      uuid.UUID          `json:"uuid" bson:"uuid"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Version   int64              `json:"version" bson:"version"`
}

func (b *BaseModel) BeforeCreate() {
	if b.ID.IsZero() {
		b.ID = primitive.NewObjectID()
	}
	if b.UUID == uuid.Nil {
		b.UUID = uuid.New()
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	if b.UpdatedAt.IsZero() {
		b.UpdatedAt = time.Now().UTC()
	}
	if b.Version == 0 {
		b.Version = 1
	}
}

func (b *BaseModel) BeforeUpdate() {
	b.UpdatedAt = time.Now().UTC()
	if b.Version == 0 {
		b.Version = 1
	} else {
		b.Version++
	}
}

func (b *BaseModel) BeforeDelete() {
	now := time.Now().UTC()
	b.DeletedAt = &now
	if b.Version == 0 {
		b.Version = 1
	} else {
		b.Version++
	}
}

func (b *BaseModel) IsDeleted() bool {
	return b.DeletedAt != nil
}





