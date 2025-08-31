package catalog

import (
	"context"
	"fmt"
	"hw11_shopql/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CataloRepoInrerface interface {
	AddCatalogWithItems(ctx context.Context, catalog model.CatalogInput) (*model.Catalog, error)
	CatalogExists(ctx context.Context, id int) (bool, error)
	AddNewCatalog(ctx context.Context, catalog model.Catalog) error
	LookupCatalog(ctx context.Context, ID int) (model.Catalog, error)
	GetItemsByCatalogID(ctx context.Context, catalogID int, limit int, offset int) ([]*model.Item, error)
}

type ItemRepoInterface interface {
	AddItem(ctx context.Context, itemInput model.ItemInput) (*model.Item, error)
}

type CatalogRepo struct {
	StMongoDB *mongo.Collection
	ItemRepoI ItemRepoInterface
}

func (CH *CatalogRepo) AddCatalogWithItems(ctx context.Context, catalog model.CatalogInput) (*model.Catalog, error) {
	var catalogItems []*model.Item
	for _, item := range catalog.Items {
		newItem, err := CH.ItemRepoI.AddItem(ctx, *item)
		catalogItems = append(catalogItems, newItem)
		if err != nil {
			return nil, err
		}
	}

	newCatalog := &model.Catalog{
		ID:       catalog.CatalogID,
		Name:     catalog.Name,
		ParentID: catalog.ParentID,
		Items:    catalogItems,
	}
	err := CH.AddNewCatalog(ctx, *newCatalog)
	if err != nil {
		return nil, err
	}
	return newCatalog, nil
}

func CreateCatalogHandler(collection *mongo.Collection, itemRepoI ItemRepoInterface) *CatalogRepo {
	return &CatalogRepo{
		StMongoDB: collection,
		ItemRepoI: itemRepoI,
	}
}

func (h *CatalogRepo) CatalogExists(ctx context.Context, id int) (bool, error) {
	filter := bson.M{"id": id}
	count, err := h.StMongoDB.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (CH *CatalogRepo) AddNewCatalog(ctx context.Context, catalog model.Catalog) error {
	if ok, _ := CH.CatalogExists(ctx, catalog.ID); !ok {
		_, err := CH.StMongoDB.InsertOne(ctx, catalog)
		if err != nil {
			return err
		}
	}
	for _, child := range catalog.Childs {
		child.ParentID = &catalog.ID
		if err := CH.AddNewCatalog(ctx, *child); err != nil {
			return err
		}
	}
	return nil
}

func InsertAllCatalogsItems(collection *mongo.Collection, item_collection *mongo.Collection, category model.Catalog) error {
	// Insert the current category
	bsonCategory, err := bson.Marshal(category)
	if err != nil {
		return fmt.Errorf("failed to marshal category: %w", err)
	}
	bsonItems, err := bson.Marshal(category.Items)
	_, err = collection.InsertOne(context.Background(), bsonCategory)
	_, err = item_collection.InsertOne(context.Background(), bsonItems)
	if err != nil {
		return fmt.Errorf("failed to insert category: %w", err)
	}

	// Recursively insert child categories
	for _, child := range category.Childs {
		if err := InsertAllCatalogsItems(collection, item_collection, *child); err != nil {
			return err
		}
	}

	return nil
}

func (CH *CatalogRepo) LookupCatalog(ctx context.Context, ID int) (model.Catalog, error) {
	filter := bson.M{
		"id": ID,
	}
	var category model.Catalog
	err := CH.StMongoDB.FindOne(ctx, filter).Decode(&category)
	if err != nil {
		fmt.Println(err)
		if err == mongo.ErrNoDocuments {
			return model.Catalog{}, nil // or return an appropriate error
		}
		return model.Catalog{}, err
	}
	return category, nil
}

func (CH *CatalogRepo) GetItemsByCatalogID(ctx context.Context, catalogID int, limit int, offset int) ([]*model.Item, error) {
	if limit <= 0 {
		limit = 3 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	filter := bson.M{
		"id": catalogID, // Assuming items have a catalog_id field
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
