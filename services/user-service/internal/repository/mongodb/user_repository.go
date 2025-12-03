package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yourorg/boilerplate/services/user-service/internal/domain"
	"github.com/yourorg/boilerplate/services/user-service/internal/repository"
	"github.com/yourorg/boilerplate/shared/errors"
)

type mongoUserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new MongoDB user repository
func NewUserRepository(db *mongo.Database) repository.UserRepository {
	return &mongoUserRepository{
		collection: db.Collection("users"),
	}
}

func (r *mongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.NewAlreadyExistsError("user", "email", user.Email)
		}
		return errors.Wrap(err, "failed to insert user")
	}
	return nil
}

func (r *mongoUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user by ID")
	}
	return &user, nil
}

func (r *mongoUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user by email")
	}
	return &user, nil
}

func (r *mongoUserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}
	if result.MatchedCount == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *mongoUserRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	if result.DeletedCount == 0 {
		return errors.ErrNotFound
	}
	return nil
}

func (r *mongoUserRepository) List(ctx context.Context, limit, offset int, search string) ([]*domain.User, int64, error) {
	filter := bson.M{}
	if search != "" {
		filter = bson.M{
			"$or": []bson.M{
				{"email": bson.M{"$regex": search, "$options": "i"}},
				{"name": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count users")
	}

	// Get paginated results
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to find users")
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, errors.Wrap(err, "failed to decode users")
	}

	return users, total, nil
}

// EnsureIndexes creates required indexes
func (r *mongoUserRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	return nil
}
