package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bocharovatd/mitm-proxy/internal/request"
	requestEntity "github.com/bocharovatd/mitm-proxy/internal/request/entity"
)

type RequestRepository struct {
	mongoCollection *mongo.Collection
}

func NewRequestRepository(mongoClient *mongo.Client) request.Repository {
	collection := mongoClient.Database("MongoBD").Collection("request")
	return &RequestRepository{mongoCollection: collection}
}

func (repository *RequestRepository) Save(req *requestEntity.HTTPRequest, resp *requestEntity.HTTPResponse, clientIP string) (string, error) {
	record := bson.M{
		"request":  req,
		"response": resp,
		"metadata": bson.M{
			"timestamp": time.Now(),
			"client_ip": clientIP,
		},
	}

	result, err := repository.mongoCollection.InsertOne(context.Background(), record)
	if err != nil {
		return "", fmt.Errorf("failed to insert request record: %v", err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return "", fmt.Errorf("failed to get inserted ID")
}

func (repository *RequestRepository) GetByID(id string) (*requestEntity.RequestRecord, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ID to ObjectID: %v", err)
	}

	filter := bson.M{"_id": objectID}
	var record requestEntity.RequestRecord

	err = repository.mongoCollection.FindOne(context.Background(), filter).Decode(&record)
	if err != nil {
		return nil, fmt.Errorf("failed to find request by ID: %v", err)
	}

	return &record, nil
}

func (repository *RequestRepository) GetAll() ([]*requestEntity.RequestRecord, error) {
	var records []*requestEntity.RequestRecord

	cursor, err := repository.mongoCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all requests: %v", err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var record requestEntity.RequestRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, fmt.Errorf("failed to decode request record: %v", err)
		}
		records = append(records, &record)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while getting all requests: %v", err)
	}

	return records, nil
}
