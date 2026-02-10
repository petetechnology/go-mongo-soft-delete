# MongoDB Soft Delete

A Go package that provides soft delete functionality for MongoDB collections.

## Installation

```bash
go get github.com/petetechnology/go-mongo-soft-delete
```

## Usage

```go
import "github.com/yourusername/mongosoftdelete"

type User struct {
    ID   primitive.ObjectID `bson:"_id"`
    Name string            `bson:"name"`

    mongo_types.SoftDeleteModel `bson:",inline"` // Embed the soft delete fields
}

func main() {
    mongoCollection := Client.DB.Collection(entities.MissingReadingsColl)

    // Initialize MongoDB collection
    softDeleteCollection := mongosoftdelete.New(mongoCollection)

    // Use soft delete features
    softDeleteResult, err := softDeleteCollection.SoftDeleteOne(ctx, bson.M{"_id": id}, deletedByID)

    // Regular query still ignores soft deleted documents
    findOneResult, err := softDeleteCollection.FindOne(ctx, bson.M{"_id": id})
}
```

## License

MIT License (or your chosen license)
