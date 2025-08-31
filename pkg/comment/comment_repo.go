package comment

import (
	"context"
	"fmt"
	"hw11_shopql/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepo struct {
	StMongoDB *mongo.Collection
}

type CommentRepoInterface interface {
	AddCommentToItem(ctx context.Context, userID int, itemID int, commentText string) (*model.Comment, error)
	AddCommentToCommnet(ctx context.Context, userID int, commentID string, commentText string) (*model.Comment, error)
}

func (CR *CommentRepo) AddCommentToItem(ctx context.Context, userID int, itemID int, commentText string) (*model.Comment, error) {
	comment := &model.Comment{
		UserID:      userID,
		ItemsID:     itemID,
		CommentText: commentText,
		Rate:        0,
	}
	_, err := CR.StMongoDB.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (CR *CommentRepo) CommentExist(ctx context.Context, commentID string) (bool, error) {
	id, _ := primitive.ObjectIDFromHex(commentID)
	filter := bson.M{
		"_id": id,
	}
	count, err := CR.StMongoDB.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (CR *CommentRepo) FindComment(ctx context.Context, commentID string) (*model.Comment, error) {
	id, _ := primitive.ObjectIDFromHex(commentID)
	filter := bson.M{
		"_id": id,
	}
	var comment *model.Comment
	err := CR.StMongoDB.FindOne(ctx, filter).Decode(&comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (CR *CommentRepo) AddCommentToCommnet(ctx context.Context, userID int, commentID string, commentText string) (*model.Comment, error) {
	exist, err := CR.CommentExist(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("comment not exist")
	}
	comm, err := CR.FindComment(ctx, commentID)
	if err != nil {
		return nil, err
	}
	comment := &model.Comment{
		UserID:      userID,
		ParentID:    &commentID,
		ItemsID:     comm.ItemsID,
		CommentText: commentText,
		Rate:        0,
	}
	_, err = CR.StMongoDB.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func CreateCommentRepo(st *mongo.Collection) *CommentRepo {
	return &CommentRepo{
		StMongoDB: st,
	}
}
