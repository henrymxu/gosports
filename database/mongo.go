package database

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/ngaut/log"
)

type MongoClient struct {
	*mongo.Client
}

type MongoDatabase struct {
	*mongo.Database
}

type MongoCollection struct {
	*mongo.Collection
}

type MongoCursor struct {
	mongo.Cursor
}

// Initialize initializes the mongo client at the provided address.
// Uses background context
func (m *MongoClient) Initialize(address string) {
	client, err := mongo.Connect(context.Background(), address)
	if err != nil {
		log.Debug("Failed to initialize database")
	}
	m.Client = client
}

// Collection returns an instance of Database with provided name databaseName
func (m *MongoClient) Database(databaseName string) Database {
	mongoDatabase := MongoDatabase{
		m.Client.Database(databaseName),
	}
	return &mongoDatabase
}

// Collection returns an instance of Collection with provided name collectionName
func (m *MongoDatabase) Collection(collectionName string) Collection {
	mongoCollection := MongoCollection{
		m.Database.Collection(collectionName),
	}
	return &mongoCollection
}

// InsertGameSnapshot inserts a snapshot into the collection
func (c *MongoCollection) InsertGameSnapshot(ctx context.Context, snapshot interface{}) {
	result, err := c.InsertOne(ctx, snapshot)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Debugf("Inserted Game Snapshot with ID: %s\n", result.InsertedID)
}

// WatchGame returns a cursor that points to a collection that will be updated when new snapshots are inserted.
func (c *MongoCollection) WatchGame(ctx context.Context) (Cursor Cursor, err error) {
	cursor, err := c.Watch(ctx, nil)
	mongoCursor := MongoCursor{
		cursor,
	}
	return &mongoCursor, err
}

func (c *MongoCursor) Close() {
	_ = c.Cursor.Close(context.Background())
}

func (c *MongoCursor) Next() bool {
	return c.Cursor.Next(context.Background())
}

func (c *MongoCursor) Decode() bson.Raw {
	bsonElement := bson.NewDocument()
	_ = c.Cursor.Decode(&bsonElement)
	return bsonElement.LookupElement("fullDocument").Value().RawDocument()
}