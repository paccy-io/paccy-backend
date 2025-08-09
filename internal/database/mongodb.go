package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/paccy-io/paccy-backend/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


type MongoDB struct {
	Client 		*mongo.Client
	Database 	*mongo.Database
	Config		*config.Config
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(cfg.GetDatabaseURI())

	clientOptions.SetMaxPoolSize(100) 
	clientOptions.SetMinPoolSize(10)  
	clientOptions.SetMaxConnIdleTime(30 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second) 
	clientOptions.SetServerSelectionTimeout(5 * time.Second) 
	clientOptions.SetRegistry(mongoRegistry) // Use the custom BSON registry for UUID support

	clientOptions.SetReadPreference(readpref.Primary())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) 
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.GetDatabaseName())

	mongodb := &MongoDB{
		Client:  client,
		Database: database,
		Config:  cfg,
	}

	log.Printf("✅ Successfully connected to MongoDB database: %s", cfg.GetDatabaseName())

	return mongodb, nil
}

func (m *MongoDB) Close() error {
	if m.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("✅ MongoDB connection closed")
	return nil
}

func (m *MongoDB) Ping(ctx context.Context) error {
	if m.Client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	return m.Client.Ping(ctx, readpref.Primary())
}

// GetDatabase returns the database instance
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.Database
}

// GetClient returns the MongoDB client
func (m *MongoDB) GetClient() *mongo.Client {
	return m.Client
}