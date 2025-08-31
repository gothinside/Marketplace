package rate

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Rate struct {
	UserID int
	ItemID int
	Rate   int
}

type RateRepo struct {
	StMongoDB *mongo.Collection
}

type RateRepoInterface interface {
	ItemsRate(ctx context.Context, itemID int) (float64, error)
	RateItem(ctx context.Context, userID, itemID, rate int) error
}

func (RR *RateRepo) RateExist(ctx context.Context, userID, itemID int) (bool, error) {
	filter := bson.M{
		"userid": userID,
		"itemid": itemID,
	}
	count, err := RR.StMongoDB.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, err
}

func (RR *RateRepo) RateItem(ctx context.Context, userID, itemID, rate int) error {

	exist, err := RR.RateExist(ctx, userID, itemID)
	if err != nil {
		return err
	}

	if !exist {
		_, err := RR.StMongoDB.InsertOne(ctx, Rate{
			UserID: userID,
			ItemID: itemID,
			Rate:   rate,
		})
		if err != nil {
			return err
		}
		return nil
	}

	filter := bson.M{
		"userid": userID,
		"itemid": itemID,
	}
	update := bson.M{
		"$set": bson.M{
			"rate": rate,
		},
	}

	_, err = RR.StMongoDB.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (RR *RateRepo) ItemsRate(ctx context.Context, itemID int) (float64, error) {
	filter := bson.D{
		{"itemid", itemID}, // Ensure the field matches your database schema
	}

	// Use the aggregation pipeline to calculate the average of rates
	pipeline := mongo.Pipeline{
		{
			{"$match", filter}, // Match documents with the specified itemID
		},
		{
			{"$group", bson.D{
				{"_id", "$itemid"},                 // Group by item_id
				{"avg", bson.D{{"$avg", "$rate"}}}, // Calculate the average of the rates
			}},
		},
	}

	cur, err := RR.StMongoDB.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx) // Ensure the cursor is closed after we're done

	var result struct {
		Avg float64 `bson:"avg"`
	}
	if cur.Next(ctx) {
		if err := cur.Decode(&result); err != nil {
			fmt.Println(err)
			return 0, err
		}
		return result.Avg, nil
	}
	if err := cur.Err(); err != nil {
		return 0, err
	}

	return result.Avg, nil
}

func CreateRateRepo(st *mongo.Collection) *RateRepo {
	return &RateRepo{StMongoDB: st}
}
