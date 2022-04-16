package repository

import (
	"context"
	"errors"
	"godas/model/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EmailVerificationRepository interface {
	Insert(context.Context, domain.EmailVerification) error
	FindByEmail(context.Context, string) (domain.EmailVerification, error)
	Update(context.Context, domain.EmailVerification) (domain.EmailVerification, error)
	Delete(context.Context, string) error
}

type EmailVerificationRepositoryImpl struct {
	collection *mongo.Collection
}

func NewEmailVerificationRepository(db *mongo.Database) EmailVerificationRepository {
	collection := db.Collection("emailVerifications")
	if _, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		panic(err)
	}

	return &EmailVerificationRepositoryImpl{
		collection: collection,
	}
}

func (repository *EmailVerificationRepositoryImpl) Insert(ctx context.Context, emailVerification domain.EmailVerification) error {
	_, err := repository.collection.InsertOne(ctx, emailVerification)
	if err != nil {
		if err, isWriteError := err.(mongo.WriteException); isWriteError && err.HasErrorCode(11000) {
			return ErrDuplicateData
		}
		return err
	}

	return nil
}

func (repository *EmailVerificationRepositoryImpl) FindByEmail(ctx context.Context, email string) (domain.EmailVerification, error) {
	emailVerification := domain.EmailVerification{}

	res := repository.collection.FindOne(ctx, bson.M{"email": email})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return emailVerification, ErrNoData
		}
		return emailVerification, err
	}

	if err := res.Decode(&emailVerification); err != nil {
		return emailVerification, err
	}

	return emailVerification, nil
}

func (repository *EmailVerificationRepositoryImpl) Update(ctx context.Context, emailVerification domain.EmailVerification) (domain.EmailVerification, error) {
	res, err := repository.collection.UpdateOne(ctx, bson.M{"email": emailVerification.Email}, bson.M{"$set": emailVerification})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return emailVerification, ErrNoData
		}
		return emailVerification, err
	}
	if res.MatchedCount < 1 {
		return emailVerification, ErrNoData
	}

	return emailVerification, nil
}

func (repository *EmailVerificationRepositoryImpl) Delete(ctx context.Context, email string) error {
	res, err := repository.collection.DeleteOne(ctx, bson.M{"email": email})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNoData
		}
		return err
	}
	if res.DeletedCount < 1 {
		return ErrNoData
	}

	return nil
}
