package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	GetCollection() *mongo.Collection
}

type service struct {
	db         *mongo.Client
	collection *mongo.Collection
}

var (
	host   = os.Getenv("DB_HOST")
	port   = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_DATABASE")
	//database = os.Getenv("DB_DATABASE")
)

func New() Service {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port)))

	if err != nil {
		log.Fatal(err)

	}
	collection := client.Database(dbName).Collection("blocked_ips")
	return &service{
		db:         client,
		collection: collection,
	}
}

// Method to return blocked IPs collection
func (s *service) GetCollection() *mongo.Collection {
	return s.collection
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Fatalf(fmt.Sprintf("db down: %v", err))
	}

	return map[string]string{
		"message": "It's healthy",
	}
}
