package repository

import (
	"context"
	"errors"
	"godas/model/domain"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type StackRepository interface {
	Insert(context.Context, domain.Stack) (domain.Stack, error)
	FindById(context.Context, string) (domain.Stack, error)
	FindByOwner(context.Context, string) ([]domain.Stack, error)
	FindAll(context.Context) ([]domain.Stack, error)
	Update(context.Context, domain.Stack) (domain.Stack, error)
	Delete(context.Context, string) error
}

type StackRepositoryImpl struct {
	collection    *mongo.Collection
	snowflakeNode *snowflake.Node
}

func NewStackRepository(db *mongo.Database, snowflakeNode *snowflake.Node) StackRepository {
	repository := new(StackRepositoryImpl)
	repository.collection = db.Collection("stacks")
	repository.snowflakeNode = snowflakeNode

	return repository
}

func (repository *StackRepositoryImpl) Insert(ctx context.Context, stack domain.Stack) (domain.Stack, error) {
	stack.ID = repository.snowflakeNode.Generate().String()

	_, err := repository.collection.InsertOne(ctx, stack)
	if err != nil {
		if err, isWriteError := err.(mongo.WriteException); isWriteError && err.HasErrorCode(11000) {
			return stack, ErrDuplicateData
		}
		return stack, err
	}

	return stack, err
}

func (repository *StackRepositoryImpl) FindById(ctx context.Context, id string) (domain.Stack, error) {
	stack := domain.Stack{}

	result := repository.collection.FindOne(ctx, bson.M{"_id": id})
	if err := result.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return stack, ErrNoData
		}
		return stack, err
	}

	if err := result.Decode(&stack); err != nil {
		return stack, err
	}

	return stack, nil
}

func (repository *StackRepositoryImpl) FindByOwner(ctx context.Context, owner string) ([]domain.Stack, error) {
	cursor, err := repository.collection.Find(ctx, bson.M{"owner": owner})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNoData
		}
		return nil, err
	}

	stacks := []domain.Stack{}
	if err := cursor.All(context.Background(), &stacks); err != nil {
		return nil, err
	}

	return stacks, nil
}

func (repository *StackRepositoryImpl) FindAll(ctx context.Context) ([]domain.Stack, error) {
	cursor, err := repository.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	stacks := []domain.Stack{}
	if err := cursor.All(ctx, &stacks); err != nil {
		return nil, err
	}

	return stacks, nil
}

func (repository *StackRepositoryImpl) Update(ctx context.Context, stack domain.Stack) (domain.Stack, error) {
	res, err := repository.collection.UpdateByID(ctx, stack.ID, bson.M{"$set": stack})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return stack, ErrNoData
		}
		return stack, err
	}
	if res.MatchedCount == 0 {
		return stack, ErrNoData
	}

	return stack, nil
}

func (repository *StackRepositoryImpl) Delete(ctx context.Context, id string) error {
	res, err := repository.collection.DeleteOne(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNoData
		}
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNoData
	}

	return nil
}
