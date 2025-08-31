package seller

import (
	"context"
	"hw11_shopql/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SellerRepoInterface interface {
	SellerExists(ctx context.Context, id int) (bool, error)
	InsertSeller(ctx context.Context, seller model.Seller) error
	LookupSellerById(ctx context.Context, id int) (*model.Seller, error)
}

type SellerRepo struct {
	StMongoDB *mongo.Collection
}

func (h *SellerRepo) SellerExists(ctx context.Context, id int) (bool, error) {
	filter := bson.M{"id": id}
	count, err := h.StMongoDB.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, err
	}
	return true, nil
}

func (SR *SellerRepo) InsertSeller(ctx context.Context, seller model.Seller) error {
	_, err := SR.StMongoDB.InsertOne(ctx, seller)
	if err != nil {
		return err
	}
	return nil
}

func (SR *SellerRepo) LookupSellerById(ctx context.Context, id int) (*model.Seller, error) {
	filter := bson.M{
		"id": id,
	}
	var seller *model.Seller
	err := SR.StMongoDB.FindOne(ctx, filter).Decode(&seller)
	if err != nil {
		return nil, err
	}
	return seller, err

}

func CreateSellersHandler(collection *mongo.Collection) *SellerRepo {
	return &SellerRepo{
		StMongoDB: collection,
	}
}
