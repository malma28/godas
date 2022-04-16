package repository

import (
	"context"
	"errors"
	"godas/model/domain"

	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Insert(context.Context, domain.User) (domain.User, error)
	FindById(context.Context, string) (domain.User, error)
	FindByEmail(context.Context, string) (domain.User, error)
	FindAll(context.Context) ([]domain.User, error)
	Update(context.Context, domain.User) (domain.User, error)
	Delete(context.Context, domain.User) error
}

type UserRepositoryImpl struct {
	collection    *mongo.Collection
	snowflakeNode *snowflake.Node
}

func NewUserRepository(db *mongo.Database, snowflakeNode *snowflake.Node) UserRepository {
	repository := new(UserRepositoryImpl)
	repository.collection = db.Collection("users")
	repository.snowflakeNode = snowflakeNode

	// Create Index
	_, err := repository.collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		panic(err)
	}

	return repository
}

func (repository *UserRepositoryImpl) Insert(ctx context.Context, user domain.User) (domain.User, error) {
	user.ID = repository.snowflakeNode.Generate().String()

	_, err := repository.collection.InsertOne(ctx, user)
	if err != nil {
		if err, isWriteException := err.(mongo.WriteException); isWriteException && err.HasErrorCode(11000) {
			return user, ErrDuplicateData
		}
		return user, err
	}

	return user, nil
}

func (repository *UserRepositoryImpl) FindById(ctx context.Context, id string) (domain.User, error) {
	user := domain.User{}

	res := repository.collection.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, ErrNoData
		}
		return user, err
	}

	err := res.Decode(&user)
	return user, err
}

func (repository *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user := domain.User{}

	res := repository.collection.FindOne(ctx, bson.M{"email": email})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, ErrNoData
		}
		return user, err
	}

	err := res.Decode(&user)
	return user, err
}

func (repository *UserRepositoryImpl) FindAll(ctx context.Context) ([]domain.User, error) {
	cur, err := repository.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	users := []domain.User{}
	err = cur.All(ctx, &users)

	return users, err
}

func (repository *UserRepositoryImpl) Update(ctx context.Context, user domain.User) (domain.User, error) {
	res, err := repository.collection.UpdateByID(ctx, user.ID, bson.M{"$set": user})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return user, ErrNoData
		}
		return user, err
	}
	if res.MatchedCount == 0 {
		return user, ErrNoData
	}

	return user, nil
}

func (repository *UserRepositoryImpl) Delete(ctx context.Context, user domain.User) error {
	res, err := repository.collection.DeleteOne(ctx, bson.M{"_id": user.ID})
	if res.DeletedCount < 1 {
		return ErrNoData
	}

	return err
}
