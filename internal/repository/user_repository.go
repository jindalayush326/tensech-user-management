package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tensechassignment/internal/model"
)

var (
	ErrDuplicateUser = errors.New("user with this email already exists for this tenant")
	ErrUserNotFound  = errors.New("user not found")
)

type BulkInsertError struct {
	Index int
	Err   error
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, tenantID, userID string) (*model.User, error)
	List(ctx context.Context, tenantID, search, status string, page, limit int64) ([]model.User, int64, error)
	CreateMany(ctx context.Context, users []model.User) (int, []BulkInsertError, error)
}

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{
		collection: db.Collection("users"),
	}
}

func (r *MongoUserRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "email", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("unique_tenant_email"),
		},
		{
			Keys: bson.D{
				{Key: "tenant_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("tenant_status_created"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *MongoUserRepository) Create(ctx context.Context, user *model.User) error {
	now := time.Now().UTC()

	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, user)

	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateUser
	}

	return err
}

func (r *MongoUserRepository) GetByID(
	ctx context.Context,
	tenantID string,
	userID string,
) (*model.User, error) {
	var user model.User

	filter := bson.M{
		"_id":       userID,
		"tenant_id": tenantID,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&user)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *MongoUserRepository) List(
	ctx context.Context,
	tenantID string,
	search string,
	status string,
	page int64,
	limit int64,
) ([]model.User, int64, error) {
	filter := bson.M{
		"tenant_id": tenantID,
	}

	if status != "" {
		filter["status"] = status
	}

	if search != "" {
		filter["$or"] = []bson.M{
			{"first_name": bson.M{
				"$regex":   search,
				"$options": "i",
			}},
			{"last_name": bson.M{
				"$regex":   search,
				"$options": "i",
			}},
			{"email": bson.M{
				"$regex":   search,
				"$options": "i",
			}},
		}
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (page - 1) * limit

	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	users := make([]model.User, 0)

	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *MongoUserRepository) CreateMany(
	ctx context.Context,
	users []model.User,
) (int, []BulkInsertError, error) {
	if len(users) == 0 {
		return 0, nil, nil
	}

	now := time.Now().UTC()
	documents := make([]interface{}, len(users))

	for i := range users {
		users[i].CreatedAt = now
		users[i].UpdatedAt = now
		documents[i] = users[i]
	}

	result, err := r.collection.InsertMany(
		ctx,
		documents,
		options.InsertMany().SetOrdered(false),
	)

	inserted := 0

	if result != nil {
		inserted = len(result.InsertedIDs)
	}

	if err == nil {
		return inserted, nil, nil
	}

	var bulkError mongo.BulkWriteException

	if !errors.As(err, &bulkError) {
		return inserted, nil, err
	}

	failures := make([]BulkInsertError, 0)

	for _, writeError := range bulkError.WriteErrors {
		if writeError.Code == 11000 {
			failures = append(failures, BulkInsertError{
				Index: writeError.Index,
				Err:   ErrDuplicateUser,
			})
			continue
		}

		failures = append(failures, BulkInsertError{
			Index: writeError.Index,
			Err:   errors.New(writeError.Message),
		})
	}

	return inserted, failures, nil
}