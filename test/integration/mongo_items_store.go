package integration

import (
	"context"
	"time"

	mongocursorpagination "github.com/qlik-oss/mongocursorpagination/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	MongoItem struct {
		ID        primitive.ObjectID `bson:"_id"`
		Name      string             `bson:"name"`
		CreatedAt time.Time          `bson:"createdAt"`
	}

	MongoStore interface {
		Create(context.Context, *MongoItem) (*MongoItem, error)
		RemoveAll(context.Context) error
		Find(ctx context.Context, query interface{}, next string, previous string, limit int64, sortAscending bool, paginatedField string, collation *options.Collation) ([]*MongoItem, mongocursorpagination.Cursor, error)
		FindBSONRaw(ctx context.Context, query interface{}, next string, previous string, limit int64, sortAscending bool, paginatedField string, collation *options.Collation) ([]bson.Raw, mongocursorpagination.Cursor, error)
	}

	mongoStore struct {
		col *mongo.Collection
	}
)

func NewMongoStore(col *mongo.Collection) MongoStore {
	return &mongoStore{
		col: col,
	}
}

// Create creates an item in the database and returns it
func (m *mongoStore) Create(ctx context.Context, c *MongoItem) (*MongoItem, error) {
	c.CreatedAt = time.Now()

	result, err := m.col.InsertOne(ctx, c)
	if err != nil {
		return nil, err
	}

	c.ID = result.InsertedID.(primitive.ObjectID)
	return c, nil
}

// Find returns paginated items from the database matching the provided query
func (m *mongoStore) Find(ctx context.Context, query interface{}, next string, previous string, limit int64, sortAscending bool, paginatedField string, collation *options.Collation) ([]*MongoItem, mongocursorpagination.Cursor, error) {
	bsonQuery := query.(bson.M)
	fp := mongocursorpagination.FindParams{
		Collection:     m.col,
		Query:          bsonQuery,
		Limit:          limit,
		SortAscending:  sortAscending,
		PaginatedField: paginatedField,
		Collation:      collation,
		Next:           next,
		Previous:       previous,
		CountTotal:     true,
	}
	var items []*MongoItem
	c, err := mongocursorpagination.Find(ctx, fp, &items)
	cursor := mongocursorpagination.Cursor{
		Previous:    c.Previous,
		Next:        c.Next,
		HasPrevious: c.HasPrevious,
		HasNext:     c.HasNext,
	}
	return items, cursor, err
}

func (m *mongoStore) FindBSONRaw(ctx context.Context, query interface{}, next string, previous string, limit int64, sortAscending bool, paginatedField string, collation *options.Collation) ([]bson.Raw, mongocursorpagination.Cursor, error) {
	bsonQuery := query.(bson.M)
	fp := mongocursorpagination.FindParams{
		Collection:     m.col,
		Query:          bsonQuery,
		Limit:          limit,
		SortAscending:  sortAscending,
		PaginatedField: paginatedField,
		Collation:      collation,
		Next:           next,
		Previous:       previous,
		CountTotal:     true,
	}
	var items []bson.Raw
	c, err := mongocursorpagination.Find(ctx, fp, &items)
	cursor := mongocursorpagination.Cursor{
		Previous:    c.Previous,
		Next:        c.Next,
		HasPrevious: c.HasPrevious,
		HasNext:     c.HasNext,
	}
	return items, cursor, err
}

func (m *mongoStore) RemoveAll(ctx context.Context) error {
	_, err := m.col.DeleteMany(ctx, bson.M{})
	return err
}
