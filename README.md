# MongoDB Soft Delete

A Go package that provides soft delete functionality for MongoDB collections using `go.mongodb.org/mongo-driver/v2`.

## Installation

```bash
go get github.com/petetechnology/go-mongo-soft-delete
```

## Usage

### Embedding the model

```go
import (
    mongosoftdelete "github.com/petetechnology/go-mongo-soft-delete"
    "go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
    ID   bson.ObjectID `bson:"_id"`
    Name string        `bson:"name"`

    mongosoftdelete.SoftDeleteModel `bson:",inline"` // Embed soft delete fields
}
```

The `SoftDeleteModel` adds the following fields to your document:

| Field       | Type            | Description                        |
|-------------|-----------------|------------------------------------|
| `deleted`   | `bool`          | Whether the document is deleted    |
| `deletedAt` | `bson.DateTime` | Timestamp of deletion              |
| `deletedBy` | `bson.ObjectID` | ID of the user who deleted it      |

### Initializing the middleware

```go
mongoCollection := client.Database("mydb").Collection("users")
coll := mongosoftdelete.New(mongoCollection)
```

### Soft deleting documents

```go
// Soft delete by filter
result, err := coll.SoftDeleteOne(ctx, bson.M{"name": "Alice"}, deletedByID)

// Soft delete by ID
result, err := coll.SoftDeleteByID(ctx, id, deletedByID)

// Soft delete multiple documents
result, err := coll.SoftDeleteMany(ctx, bson.M{"role": "guest"}, deletedByID)
```

### Querying (soft-deleted documents are automatically excluded)

```go
// Find one
result := coll.FindOne(ctx, bson.M{"_id": id})

// Find many
cursor, err := coll.Find(ctx, bson.M{"role": "admin"})

// Count
count, err := coll.CountDocuments(ctx, bson.M{})

// Aggregate
cursor, err := coll.Aggregate(ctx, bson.A{
    bson.D{{Key: "$group", Value: bson.M{"_id": "$role", "count": bson.M{"$sum": 1}}}},
})
```

### Updating documents

```go
// Update one by filter
result, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"name": "Bob"}})

// Update one by ID
result, err := coll.UpdateByID(ctx, id, bson.M{"$set": bson.M{"name": "Bob"}})

// Update many
result, err := coll.UpdateMany(ctx, bson.M{"role": "guest"}, bson.M{"$set": bson.M{"active": false}})

// Find one and update
result := coll.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"name": "Bob"}})
```

### Inserting documents

```go
// Insert one
result, err := coll.InsertOne(ctx, user)

// Insert many
result, err := coll.InsertMany(ctx, []any{user1, user2})
```

## Interface

The package exposes `ISoftDeleteMiddleware` for use in mocks and dependency injection:

```go
type MyRepo struct {
    coll mongosoftdelete.ISoftDeleteMiddleware
}
```

## License

MIT
