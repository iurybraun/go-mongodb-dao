/*
 * Copyright © 2016-2022 Iury Braun
 * Copyright © 2017-2022 BRAUN TECH
 *
 * Refatorado: correções de context leak, bug de referência em cursor,
 * os.Exit/log.Fatal removidos, cursor.Close corrigido, URL via env,
 * init() substituído por Connect(), Upsert e InsertMany adicionados,
 * SetDatabase/GetDatabase para configuração global do banco de dados.
 */

package go_mongodb_dao

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const defaultTimeout = 10 * time.Second
const queryTimeout = 120 * time.Second

// Dao struct holds the target database and collection names.
// If Database is empty, the value set via SetDatabase() is used.
type Dao struct {
	Database   string
	Collection string
}

// database resolves the effective database name: uses Dao.Database when
// explicitly set, otherwise falls back to the global dbName configured
// via SetDatabase().
func (m *Dao) database() string {
	if m.Database != "" {
		return m.Database
	}
	return dbName
}

var (
	client *mongo.Client
	dbName string
)

// SetDatabase defines the default database name used by all DAO operations
// when Dao.Database is empty. Call this once during application startup.
//
//	dao.SetDatabase(cfg_ini.LoadKey_string("database", "dbname"))
func SetDatabase(name string) {
	dbName = name
}

// GetDatabase returns the currently configured default database name.
func GetDatabase() string {
	return dbName
}

// getMongoURL returns the MongoDB connection URL from the environment
// variable MONGO_URL, falling back to localhost if not set.
func getMongoURL() string {
	if url := os.Getenv("MONGO_URL"); url != "" {
		return url
	}
	return "mongodb://localhost:27017"
}

// Connect initialises the global MongoDB client. Call this once at
// application startup instead of relying on init().
//
//	if err := go_mongodb_dao.Connect(""); err != nil {
//	    log.Fatal(err)
//	}
func Connect(mongoURI string) error {
	if mongoURI == "" {
		mongoURI = getMongoURL()
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return err
	}

	if err = conn.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	client = conn
	return nil
}

// Disconnect gracefully closes the MongoDB connection.
func Disconnect() error {
	if client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return client.Disconnect(ctx)
}

// -----------------------------------------------------------------
// INSERT
// -----------------------------------------------------------------

// Insert adds a single document to the collection and returns its ID.
func (m *Dao) Insert(doc interface{}) (interface{}, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}

// InsertMany adds multiple documents to the collection in a single
// round-trip and returns their inserted IDs.
func (m *Dao) InsertMany(docs []interface{}) ([]interface{}, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	result, err := collection.InsertMany(context.TODO(), docs)
	if err != nil {
		return nil, err
	}

	return result.InsertedIDs, nil
}

// -----------------------------------------------------------------
// FIND
// -----------------------------------------------------------------

// FindByID retrieves a single document by its ObjectID hex string.
func (m *Dao) FindByID(_id string) (map[string]interface{}, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	objID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		return nil, fmt.Errorf("FindByID — invalid ObjectID %q: %w", _id, err)
	}

	doc := make(map[string]interface{})
	findOne := collection.FindOne(context.TODO(), bson.M{"_id": objID})
	if err := findOne.Err(); err != nil {
		return nil, err
	}

	if err = findOne.Decode(&doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// FindOneWithFilters retrieves the first document matching the given query.
func (m *Dao) FindOneWithFilters(qry map[string]interface{}) (map[string]interface{}, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	findOne := collection.FindOne(context.TODO(), qry)
	if err := findOne.Err(); err != nil {
		return nil, err
	}

	doc := make(map[string]interface{})
	if err := findOne.Decode(&doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// FindAll returns every document in the collection.
func (m *Dao) FindAll() ([]map[string]interface{}, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	docs := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		// Allocate a new map per iteration to avoid all entries
		// pointing to the same underlying map (reference bug).
		n := make(map[string]interface{})
		if err := cursor.Decode(&n); err != nil {
			return nil, err
		}
		docs = append(docs, n)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

// FindAllWithFilters returns documents matching qry with optional
// sorting, limit, and cursor-based pagination (after / before ObjectID
// hex strings).
func (m *Dao) FindAllWithFilters(
	qry map[string]interface{},
	sort map[string]int,
	limit int,
	after, before string,
) ([]map[string]interface{}, error) {

	collection := client.Database(m.database()).Collection(m.Collection)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	findOptions := options.Find()
	if sort != nil {
		findOptions.SetSort(sort)
	}
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}

	// Cursor-based pagination: mutating the caller's map is intentional
	// (consistent with prior behaviour), but we do it safely.
	if after != "" {
		objID, err := primitive.ObjectIDFromHex(after)
		if err != nil {
			return nil, fmt.Errorf("FindAllWithFilters — invalid 'after' ObjectID %q: %w", after, err)
		}
		qry["_id"] = bson.M{"$gt": objID}
	}

	if before != "" {
		objID, err := primitive.ObjectIDFromHex(before)
		if err != nil {
			return nil, fmt.Errorf("FindAllWithFilters — invalid 'before' ObjectID %q: %w", before, err)
		}
		qry["_id"] = bson.M{"$lt": objID}
	}

	cursor, err := collection.Find(ctx, qry, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	docs := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		n := make(map[string]interface{})
		if err := cursor.Decode(&n); err != nil {
			return nil, err
		}
		docs = append(docs, n)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

// -----------------------------------------------------------------
// AGGREGATE
// -----------------------------------------------------------------

// Aggregate executes a pipeline and returns the resulting documents.
func (m *Dao) Aggregate(pipeline interface{}) ([]map[string]interface{}, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	docs := make([]map[string]interface{}, 0)
	for cursor.Next(ctx) {
		n := make(map[string]interface{})
		if err := cursor.Decode(&n); err != nil {
			return nil, err
		}
		docs = append(docs, n)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

// -----------------------------------------------------------------
// UPDATE
// -----------------------------------------------------------------

// Update replaces the fields of a single document (identified by _id)
// using $set.
func (m *Dao) Update(_id string, doc interface{}) (int64, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	objID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		return -1, fmt.Errorf("Update — invalid ObjectID %q: %w", _id, err)
	}

	result, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$set": doc},
	)
	if err != nil {
		return -1, err
	}

	return result.ModifiedCount, nil
}

// UpdateWithFilters applies doc (which must already contain a MongoDB
// update operator such as $set, $unset, etc.) to all documents that
// match qry.
func (m *Dao) UpdateWithFilters(qry map[string]interface{}, doc interface{}) (int64, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	result, err := collection.UpdateMany(context.TODO(), qry, doc)
	if err != nil {
		return -1, err
	}

	return result.ModifiedCount, nil
}

// UpdateManyWithFilters applies $set with doc to all documents matching
// qry.
func (m *Dao) UpdateManyWithFilters(qry map[string]interface{}, doc interface{}) (int64, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	result, err := collection.UpdateMany(
		context.TODO(),
		qry,
		bson.M{"$set": doc},
	)
	if err != nil {
		return -1, err
	}

	return result.ModifiedCount, nil
}

// Upsert updates the first document matching qry using $set, or inserts
// a new document when no match is found. Returns the number of documents
// modified or inserted.
func (m *Dao) Upsert(qry map[string]interface{}, doc interface{}) (int64, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(
		context.TODO(),
		qry,
		bson.M{"$set": doc},
		opts,
	)
	if err != nil {
		return -1, err
	}

	return result.ModifiedCount + result.UpsertedCount, nil
}

// -----------------------------------------------------------------
// DELETE
// -----------------------------------------------------------------

// Delete removes the document with the given ObjectID hex string.
func (m *Dao) Delete(_id string) (int64, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	objID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		return -1, fmt.Errorf("Delete — invalid ObjectID %q: %w", _id, err)
	}

	result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return -1, err
	}

	return result.DeletedCount, nil
}

// DeleteAll removes every document matching qry from the collection.
func (m *Dao) DeleteAll(qry map[string]interface{}) (int64, error) {
	collection := client.Database(m.database()).Collection(m.Collection)

	result, err := collection.DeleteMany(context.TODO(), qry)
	if err != nil {
		return -1, err
	}

	return result.DeletedCount, nil
}

// -----------------------------------------------------------------
// MISC
// -----------------------------------------------------------------

// Stats returns collection-level statistics via the collStats command.
func (m *Dao) Stats() (map[string]interface{}, error) {
	db := client.Database(m.database())

	result := db.RunCommand(context.Background(), bson.M{"collStats": m.Collection})

	var document bson.M
	if err := result.Decode(&document); err != nil {
		return nil, err
	}

	return document, nil
}
