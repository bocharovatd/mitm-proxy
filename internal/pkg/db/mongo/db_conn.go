package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoURL = "mongodb://mongo:27017"
)

func New() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, fmt.Errorf("error creating and connecting MongoDB client: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("error pinging MongoDB: %v", err)
	}

	return client, nil
}
