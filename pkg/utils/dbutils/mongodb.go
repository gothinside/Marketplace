package dbutils

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDbHandler struct {
	Db *mongo.Database
}

func (MDH *MongoDbHandler) DeleteFromCollection(collection string) error {
	client := MDH.Db.Collection(collection)
	filter := bson.M{}
	_, err := client.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}
