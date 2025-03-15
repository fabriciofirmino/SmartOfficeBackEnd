package utils

import (
	"apiBackEnd/config"
	"context"
	"fmt"
	"time"
)

// SaveToMongo salva um documento em uma coleção do MongoDB
func SaveToMongo(collectionName string, data interface{}) error {
	if config.MongoDB == nil {
		return fmt.Errorf("MongoDB não está inicializado")
	}

	collection := config.MongoDB.Database("api_logs").Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return fmt.Errorf("erro ao salvar no MongoDB: %v", err)
	}

	return nil
}
