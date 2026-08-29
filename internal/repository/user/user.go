package user

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	userdomain "user-management/internal/domain/user"
)

// userDoc is the persistence shape of userdomain.User. It keeps the ObjectID out
// of the domain layer.
type userDoc struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
}

func (d userDoc) toDomain() *userdomain.User {
	return &userdomain.User{
		ID:        d.ID.Hex(),
		Name:      d.Name,
		Email:     d.Email,
		Password:  d.Password,
		CreatedAt: d.CreatedAt,
	}
}

// Mongo is the MongoDB implementation of Repository.
type Mongo struct {
	coll *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{coll: db.Collection("users")}
}

func (r *Mongo) Create(ctx context.Context, u *userdomain.User) error {
	u.CreatedAt = time.Now()
	res, err := r.coll.InsertOne(ctx, userDoc{
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
	})
	if err != nil {
		return err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("unexpected inserted id type")
	}
	u.ID = oid.Hex()
	return nil
}

func (r *Mongo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	var d userDoc
	if err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&d); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, userdomain.ErrNotFound
		}
		return nil, err
	}
	return d.toDomain(), nil
}

func (r *Mongo) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, userdomain.ErrNotFound
	}
	var d userDoc
	if err := r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&d); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, userdomain.ErrNotFound
		}
		return nil, err
	}
	return d.toDomain(), nil
}

func (r *Mongo) List(ctx context.Context) ([]*userdomain.User, error) {
	cur, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	users := []*userdomain.User{}
	for cur.Next(ctx) {
		var d userDoc
		if err := cur.Decode(&d); err != nil {
			return nil, err
		}
		users = append(users, d.toDomain())
	}
	return users, cur.Err()
}

func (r *Mongo) Update(ctx context.Context, u *userdomain.User) error {
	oid, err := primitive.ObjectIDFromHex(u.ID)
	if err != nil {
		return userdomain.ErrNotFound
	}
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"name": u.Name, "email": u.Email}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return userdomain.ErrNotFound
	}
	return nil
}

func (r *Mongo) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return userdomain.ErrNotFound
	}
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return userdomain.ErrNotFound
	}
	return nil
}

func (r *Mongo) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}
