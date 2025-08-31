package order

import (
	"context"
	"hw11_shopql/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CartRepoInterface interface {
	GetCartItems(ctx context.Context, UserID int) ([]*model.CartItem, error)
}

type ItemRepoInterface interface {
	UpdateItemQuantity(ctx context.Context, itemID, newQuantity int) error
}

type OrderRepo struct {
	St        *mongo.Collection
	CartRepoI CartRepoInterface
	ItemRepoI ItemRepoInterface
	count     int
}

type OrderRepoInterface interface {
	CreateOrder(ctx context.Context, userID int) (*model.Order, error)
	UsersOrders(ctx context.Context, userID int) ([]*model.Order, error)
}

func (OR *OrderRepo) CreateOrder(ctx context.Context, userID int) (*model.Order, error) {
	items, err := OR.CartRepoI.GetCartItems(ctx, userID)
	order := &model.Order{}
	order.UserID = userID
	order.OrderID = OR.count
	OR.count++

	if err != nil {
		return nil, err
	}

	for _, item := range items {
		err = OR.ItemRepoI.UpdateItemQuantity(ctx, item.Item.ID, item.Item.InStock-item.Quantity)
		order.Items = append(order.Items, item)
		if err != nil {
			return nil, err
		}
	}

	order.Status = "Order created"
	_, err = OR.St.InsertOne(ctx, order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (OR *OrderRepo) UsersOrders(ctx context.Context, userID int) ([]*model.Order, error) {
	filter := bson.M{"userid": userID}
	cur, err := OR.St.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx) // Ensure the cursor is closed after we're done

	var UsersOrders []*model.Order
	for cur.Next(ctx) {
		var order *model.Order
		if err := cur.Decode(&order); err != nil {
			return nil, err
		}
		UsersOrders = append(UsersOrders, order)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	return UsersOrders, nil
}
func CreateOrderRepo(St *mongo.Collection, cartRepoI CartRepoInterface, itemRepoI ItemRepoInterface) *OrderRepo {
	return &OrderRepo{
		St:        St,
		CartRepoI: cartRepoI,
		ItemRepoI: itemRepoI,
		count:     1,
	}
}
