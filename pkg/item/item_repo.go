package item

import (
	"context"
	"fmt"
	"hw11_shopql/graph/model"
	"hw11_shopql/pkg/comment"
	"hw11_shopql/pkg/rate"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ItemRepoInterface interface {
	AddItem(ctx context.Context, itemInput model.ItemInput) (*model.Item, error)
	UpdateItemQuantity(ctx context.Context, itemID, newQuantity int) error
	AddCommentToCommnet(ctx context.Context, userID int, commentID string, commentText string) (*model.Comment, error)
	AddComment(ctx context.Context, userID, itemID int, commentText string) (*model.Comment, error)
	ItemsRate(ctx context.Context, itemID int) (float64, error)
	RateItem(ctx context.Context, userID, itemID, rate int) (*model.Item, error)
	GetItemByID(ctx context.Context, id int) (*model.Item, error)
	InStockByQuantity(quantity int) string
	InsertCatalogsItems(ctx context.Context, catalog model.Catalog) error
	ItemExists(ctx context.Context, id int) (bool, error)
	GetItemsByCatalogID(ctx context.Context, catalogID int, limit int, offset int) ([]*model.Item, error)
	GetItemsBySellerID(ctx context.Context, seller_id int, limit *int, offset *int) ([]*model.Item, error)
}

type ItemRepo struct {
	StMongoDB   *mongo.Collection
	RateRepo    rate.RateRepoInterface
	CommentRepo comment.CommentRepoInterface
}

func (IH *ItemRepo) AddItem(ctx context.Context, itemInput model.ItemInput) (*model.Item, error) {
	if itemInput.InStock < 0 {
		return nil, fmt.Errorf("instock can't be less then 0")
	}
	if ok, _ := IH.ItemExists(ctx, itemInput.ItemID); ok {
		return nil, fmt.Errorf("item already exist")
	}
	item := &model.Item{
		ID:          itemInput.ItemID,
		Name:        itemInput.Name,
		SellerID:    itemInput.SellerID,
		InStock:     itemInput.InStock,
		Rate:        0,
		InStockText: IH.InStockByQuantity(itemInput.InStock),
		CatalogID:   itemInput.CatalogID,
	}
	_, err := IH.StMongoDB.InsertOne(ctx, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (IH *ItemRepo) UpdateItemQuantity(ctx context.Context, itemID, newQuantity int) error {
	filter := bson.M{
		"id": itemID,
	}
	update := bson.M{
		"$set": bson.M{
			"in_stock": newQuantity,
		},
	}
	_, err := IH.StMongoDB.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return err
}

func (IH *ItemRepo) AddCommentToCommnet(ctx context.Context, userID int, commentID string, commentText string) (*model.Comment, error) {
	comment, err := IH.CommentRepo.AddCommentToCommnet(ctx, userID, commentID, commentText)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (IH *ItemRepo) AddComment(ctx context.Context, userID, itemID int, commentText string) (*model.Comment, error) {
	comment, err := IH.CommentRepo.AddCommentToItem(ctx, userID, itemID, commentText)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (IH *ItemRepo) ItemsRate(ctx context.Context, itemID int) (float64, error) {
	rate, err := IH.RateRepo.ItemsRate(ctx, itemID)
	if err != nil {
		return 0, err
	}
	return rate, nil
}

func (IH *ItemRepo) RateItem(ctx context.Context, userID, itemID, rate int) (*model.Item, error) {
	err := IH.RateRepo.RateItem(ctx, userID, itemID, rate)
	if err != nil {
		return nil, err
	}
	item, err := IH.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (IH *ItemRepo) GetItemByID(ctx context.Context, id int) (*model.Item, error) {
	filter := bson.M{"id": id}
	var item *model.Item
	err := IH.StMongoDB.FindOne(ctx, filter).Decode(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (IH *ItemRepo) InStockByQuantity(quantity int) string {
	if quantity <= 1 {
		return "мало"
	} else if quantity > 1 && quantity <= 3 {
		return "хватает"
	} else {
		return "много"
	}
}

func (IH *ItemRepo) InsertCatalogsItems(ctx context.Context, catalog model.Catalog) error {
	for _, item := range catalog.Items {
		item.CatalogID = catalog.ID
		if item.InStock == 1 {
			item.InStockText = "мало"
		} else if item.InStock > 1 && item.InStock <= 3 {
			item.InStockText = "хватает"
		} else {
			item.InStockText = "много"
		}
		if ok, _ := IH.ItemExists(ctx, item.ID); !ok {
			_, err := IH.StMongoDB.InsertOne(ctx, item)
			if err != nil {
				return err
			}
		}
	}

	for _, child := range catalog.Childs {

		if err := IH.InsertCatalogsItems(ctx, *child); err != nil {
			return err
		}
	}
	return nil
}

func (IH *ItemRepo) ItemExists(ctx context.Context, id int) (bool, error) {
	filter := bson.M{"id": id}
	count, err := IH.StMongoDB.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil

}

func (CH *ItemRepo) GetItemsByCatalogID(ctx context.Context, catalogID int, limit int, offset int) ([]*model.Item, error) {
	if limit <= 0 {
		limit = 3 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	filter := bson.M{
		"catalogid": catalogID, // Assuming items have a catalog_id field
	}

	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := CH.StMongoDB.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find items: %w", err)
	}
	defer cursor.Close(ctx)

	var items []*model.Item
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("failed to decode items: %w", err)
	}

	return items, nil
}

func (CH *ItemRepo) GetItemsBySellerID(ctx context.Context, seller_id int, limit *int, offset *int) ([]*model.Item, error) {
	filter := bson.M{
		"sellerid": seller_id, // Assuming items have a catalog_id field
	}
	cursor, err := CH.StMongoDB.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find items: %w", err)
	}
	defer cursor.Close(ctx)
	var items []*model.Item
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("failed to decode items: %w", err)
	}

	return items, nil
}

func CreateItemsHandler(collection *mongo.Collection, rateRepoI rate.RateRepoInterface, commentRepoI comment.CommentRepoInterface) *ItemRepo {
	return &ItemRepo{
		StMongoDB:   collection,
		RateRepo:    rateRepoI,
		CommentRepo: commentRepoI,
	}
}

// func InsertItems(ctx context.Context, items []*model.Item){

// }
