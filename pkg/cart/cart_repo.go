package cart

import (
	"context"
	"fmt"

	"hw11_shopql/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ItemRepoInterface interface {
	GetItemByID(ctx context.Context, id int) (*model.Item, error)
	InStockByQuantity(quantity int) string
}

type Cart struct {
	User_id  int
	Item_id  int
	Quantity int
}

type CartRepoInterface interface {
	CartExist(ctx context.Context, UserID int, ItemID int) (bool, error)
	GetCartsItem(ctx context.Context, UserID int, ItemID int) (*Cart, error)
	AddItem(ctx context.Context, cart *model.CartInput, UserID int) error
	RemoveFromCartItem(ctx context.Context, cart *model.CartInput, UserID int) error
	GetCartItems(ctx context.Context, UserID int) ([]*model.CartItem, error)
}

type CartRepo struct {
	St          *mongo.Collection
	ItemStorage ItemRepoInterface
}

func (CR *CartRepo) CartExist(ctx context.Context, UserID int, ItemID int) (bool, error) {
	filter := bson.M{"user_id": UserID, "item_id": ItemID}
	count, err := CR.St.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (CR *CartRepo) GetCartsItem(ctx context.Context, UserID int, ItemID int) (*Cart, error) {
	filter := bson.M{"user_id": UserID, "item_id": ItemID}
	cart := &Cart{}
	err := CR.St.FindOne(ctx, filter).Decode(&cart)
	if err != nil {
		return nil, err
	}
	return cart, nil

}

func (CR *CartRepo) AddItem(ctx context.Context, cart *model.CartInput, UserID int) error {
	exist, err := CR.CartExist(ctx, UserID, cart.ItemID)
	if err != nil {
		return err
	}
	item, err := CR.ItemStorage.GetItemByID(ctx, cart.ItemID)
	if err != nil {
		return err
	}
	if !exist {
		if item.InStock < cart.Quantity {
			return fmt.Errorf("not enough quantity")
		}
		cart := Cart{
			User_id:  UserID,
			Item_id:  cart.ItemID,
			Quantity: cart.Quantity,
		}
		_, err = CR.St.InsertOne(ctx, cart)
		if err != nil {
			return err
		}
		return nil
	} else {
		CartItem, err := CR.GetCartsItem(ctx, UserID, cart.ItemID)
		if err != nil {
			return err
		}
		if cart.Quantity+CartItem.Quantity > item.InStock {
			return fmt.Errorf("not enough quantity")
		}

		newQuantity := CartItem.Quantity + cart.Quantity
		filter := bson.M{
			"user_id": UserID,
			"item_id": cart.ItemID,
		}
		update := bson.M{
			"$set": bson.M{
				"quantity": newQuantity,
			},
		}
		_, err = CR.St.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (CR *CartRepo) RemoveFromCartItem(ctx context.Context, cart *model.CartInput, UserID int) error {
	exist, err := CR.CartExist(ctx, UserID, cart.ItemID)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("cart in no exist")
	}
	item, err := CR.ItemStorage.GetItemByID(ctx, cart.ItemID)
	if err != nil {
		return err
	}
	CartItem, err := CR.GetCartsItem(ctx, UserID, cart.ItemID)
	if err != nil {
		return err
	}
	if CartItem.Quantity-cart.Quantity <= 0 {
		filter := bson.M{
			"user_id": UserID,
			"item_id": cart.ItemID,
		}
		_, err = CR.St.DeleteOne(ctx, filter)
		if err != nil {
			return err
		}
		return nil
	} else {
		CartItem, err := CR.GetCartsItem(ctx, UserID, cart.ItemID)
		if err != nil {
			return err
		}
		if cart.Quantity+CartItem.Quantity > item.InStock {
			return fmt.Errorf("not enough quantity")
		}

		newQuantity := CartItem.Quantity - cart.Quantity
		filter := bson.M{
			"user_id": UserID,
			"item_id": cart.ItemID,
		}
		update := bson.M{
			"$set": bson.M{
				"quantity": newQuantity,
			},
		}
		CR.St.UpdateOne(ctx, filter, update)
	}
	return nil
}

func (CR *CartRepo) GetCartItems(ctx context.Context, UserID int) ([]*model.CartItem, error) {
	filter := bson.M{"user_id": UserID}
	cur, err := CR.St.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx) // Ensure the cursor is closed after we're done

	var cartItems []*model.CartItem
	for cur.Next(ctx) {
		var cart Cart
		if err := cur.Decode(&cart); err != nil {
			return nil, err
		}
		item, _ := CR.ItemStorage.GetItemByID(ctx, cart.Item_id)
		item.InStockText = CR.ItemStorage.InStockByQuantity(item.InStock - cart.Quantity)
		cartItems = append(cartItems, &model.CartItem{Quantity: cart.Quantity, Item: item})
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return cartItems, nil
}

func CreateCartRepo(St *mongo.Collection, itemRepo ItemRepoInterface) *CartRepo {
	return &CartRepo{
		St:          St,
		ItemStorage: itemRepo,
	}
}
