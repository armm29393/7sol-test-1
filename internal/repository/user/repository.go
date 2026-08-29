package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"user-management/internal/domain"
)

type UserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	u.CreatedAt = time.Now()
	res, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrUserNotFound
	}
	return &u, err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	var u domain.User
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, domain.ErrUserNotFound
	}
	return &u, err
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	cur, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var users []*domain.User
	for cur.Next(ctx) {
		var u domain.User
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, id, name, email string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"name": name, "email": email}},
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	oid, _ := primitive.ObjectIDFromHex(id)
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

func Connect(uri, dbName string) (*mongo.Database, func(context.Context) error, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, nil, err
	}
	return client.Database(dbName), client.Disconnect, nil
}
