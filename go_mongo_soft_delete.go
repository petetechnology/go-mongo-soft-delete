package mongosoftdelete

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SoftDeleteModel contains common fields for soft deletion
type SoftDeleteModel struct {
	Deleted   bool               `bson:"deleted,omitempty"`
	DeletedAt primitive.DateTime `bson:"deletedAt,omitempty"`
	DeletedBy primitive.ObjectID `bson:"deletedBy,omitempty"`
}

// SoftDeleteMiddleware adds soft delete filter to all queries
type SoftDeleteMiddleware struct {
	*mongo.Collection
}

// ISoftDeleteMiddleware defines the interface for soft deletion operations
type ISoftDeleteMiddleware interface {
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	SoftDeleteOne(ctx context.Context, filter interface{}, deletedBy primitive.ObjectID) (*mongo.UpdateResult, error)
	SoftDeleteMany(ctx context.Context, filter interface{}, deletedBy primitive.ObjectID) (*mongo.UpdateResult, error)
	SoftDeleteByID(ctx context.Context, id primitive.ObjectID, deletedBy primitive.ObjectID) (*mongo.UpdateResult, error)
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error)
	UpdateByID(ctx context.Context, id primitive.ObjectID, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	InsertOne(ctx context.Context, create interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	InsertMany(ctx context.Context, creates []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)

	Indexes() mongo.IndexView
}

func New(coll *mongo.Collection) *SoftDeleteMiddleware {
	return &SoftDeleteMiddleware{Collection: coll}
}

// Find adds a soft delete filter to the query. It ensures that only documents with deleted=false are returned.
func (m *SoftDeleteMiddleware) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	filter = m.addSoftDeleteFilter(filter)
	return m.Collection.Find(ctx, filter, opts...)
}

// FindOne adds a soft delete filter to the query. It ensures that only documents with deleted=false are returned.
func (m *SoftDeleteMiddleware) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	filter = m.addSoftDeleteFilter(filter)
	return m.Collection.FindOne(ctx, filter, opts...)
}

// SoftDeleteOne performs a soft delete operation
func (m *SoftDeleteMiddleware) SoftDeleteOne(ctx context.Context, filter interface{}, deletedBy primitive.ObjectID) (*mongo.UpdateResult, error) {
	update := m.createSoftDeleteUpdate(deletedBy)
	return m.Collection.UpdateOne(ctx, filter, update)
}

// SoftDeleteMany performs a soft delete operation on multiple documents
func (m *SoftDeleteMiddleware) SoftDeleteMany(ctx context.Context, filter interface{}, deletedBy primitive.ObjectID) (*mongo.UpdateResult, error) {
	update := m.createSoftDeleteUpdate(deletedBy)
	return m.Collection.UpdateMany(ctx, filter, update)
}

// SoftDeleteByID performs a soft delete operation on a document by its ID
func (m *SoftDeleteMiddleware) SoftDeleteByID(ctx context.Context, id primitive.ObjectID, deletedBy primitive.ObjectID) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": id}
	update := m.createSoftDeleteUpdate(deletedBy)
	return m.Collection.UpdateOne(ctx, filter, update)
}

// Add the new methods to SoftDeleteMiddleware
// Aggregate adds a soft delete filter to the aggregation pipeline
func (m *SoftDeleteMiddleware) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	// Convert pipeline to array if it's not already
	pipelineArray, ok := pipeline.([]interface{})
	if !ok {
		switch p := pipeline.(type) {
		case bson.D:
			pipelineArray = []interface{}{p}
		case bson.A:
			pipelineArray = p
		default:
			return nil, fmt.Errorf("pipeline must be []interface{}, bson.D or bson.A, got %T", pipeline)
		}
	}

	// Add $match stage with soft delete filter at the beginning of pipeline
	softDeleteMatch := bson.D{{
		Key: "$match",
		Value: bson.M{
			"deleted": bson.M{"$ne": true},
		},
	}}

	newPipeline := append([]interface{}{softDeleteMatch}, pipelineArray...)

	return m.Collection.Aggregate(ctx, newPipeline, opts...)
}

// UpdateByID updates a single document by ID
func (m *SoftDeleteMiddleware) UpdateByID(ctx context.Context, id primitive.ObjectID, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter := bson.M{
		"_id":     id,
		"deleted": bson.M{"$ne": true},
	}

	return m.Collection.UpdateOne(ctx, filter, update, opts...)
}

// You might also want to add a convenience method for updating non-deleted documents
func (m *SoftDeleteMiddleware) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter = m.addSoftDeleteFilter(filter)
	return m.Collection.UpdateOne(ctx, filter, update, opts...)
}

// And a method for updating many documents
func (m *SoftDeleteMiddleware) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	filter = m.addSoftDeleteFilter(filter)
	return m.Collection.UpdateMany(ctx, filter, update, opts...)
}

// FindOneAndUpdate adds a soft delete filter to the query and performs a find-and-update operation
func (m *SoftDeleteMiddleware) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	filter = m.addSoftDeleteFilter(filter)
	return m.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
}

// InsertOne inserts a single document into the collection
func (m *SoftDeleteMiddleware) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return m.Collection.InsertOne(ctx, document, opts...)
}

// InsertMany inserts multiple documents into the collection
func (m *SoftDeleteMiddleware) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return m.Collection.InsertMany(ctx, documents, opts...)
}

// addSoftDeleteFilter adds a soft delete filter to the query. It ensures that only documents with deleted=false are returned.
func (m *SoftDeleteMiddleware) addSoftDeleteFilter(filter interface{}) interface{} {
	if filter == nil {
		return bson.M{"deleted": bson.M{"$ne": true}}
	}

	return bson.M{
		"$and": []interface{}{
			filter,

			// Handling when the field does not exist or is false.
			bson.M{"deleted": bson.M{"$ne": true}},
		},
	}
}

// createSoftDeleteUpdate creates the update payload for a soft delete operation. It sets the deleted field to true and adds the deletedAt timestamp. If deletedBy is provided, it also sets the deletedBy field.
func (m *SoftDeleteMiddleware) createSoftDeleteUpdate(deletedBy primitive.ObjectID) bson.M {
	softDeletionPayload := bson.M{
		"deleted":   true,
		"deletedAt": primitive.NewDateTimeFromTime(time.Now()),
	}

	// handling when deletedBy is provided
	if !deletedBy.IsZero() {
		softDeletionPayload["deletedBy"] = deletedBy
	}

	return bson.M{
		"$set": softDeletionPayload,
	}
}
