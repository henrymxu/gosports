package database

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
)

type client interface {
	Initialize(address string)
	Database(database string) Database
}

type Database interface {
	Collection(collection string) Collection
}

type Collection interface {
	InsertGameSnapshot(ctx context.Context, snapshot interface{})
	WatchGame(ctx context.Context) (cursor Cursor, err error)
}

type Cursor interface {
	Close()
	Next() bool
	Decode() bson.Raw
}



