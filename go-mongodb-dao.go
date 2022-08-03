/*
 * Copyright © 2016-2022 Iury Braun
 * Copyright © 2017-2022 Neo7i
 * 
 * Alt:  Id bson.ObjectId  ==>  id interface{}
 * 
 * POOL:
 *  https://stackoverflow.com/questions/57998402/how-can-i-convert-my-mgo-sessions-to-mongo-go-driver-clients-using-connection-po
 *  https://kb.objectrocket.com/mongo-db/how-to-get-mongodb-documents-using-golang-446
 * 
 *  https://github.com/FrevChuso/mongo-crud-base
 *  https://github.com/Pokervarino27/crud_go_mongodb
 *  https://github.com/cloudgate313/go-mongodb-crud
 * 
 * PAGINATION:
 *  https://github.com/carlosstrand/graphql-pagination-go
 *  https://shopify.dev/concepts/graphql
 *  https://medium.com/@mattmazzola/graphql-pagination-implementation-8604f77fb254
 *  https://stackoverflow.com/questions/51179588/how-to-sort-and-limit-results-in-mongodb
 */

package go_mongodb_dao

import (
    "log"
    "fmt"
    "time"
    "context"
    "os"
    //"net/url"
    
    "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	//"go.mongodb.org/mongo-driver/event"
	//"go.mongodb.org/mongo-driver/internal/testutil/helpers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
	//"go.mongodb.org/mongo-driver/mongo/readconcern"
	//"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Dao struct ...
type Dao struct {
	Database   string
	Collection string
}


var (
    client     *mongo.Client
    mongoURL = "mongodb://localhost:27017"
)

// Connect with mongo server
func init() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    conn, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatalln(err)
	}
    
    if err = conn.Ping(ctx, readpref.Primary()); err != nil {
        log.Fatal(err)
    }
    
    client = conn
}


func (m *Dao) Insert(doc interface{}) (interface{}, error) {
    var collection = client.Database(m.Database).Collection(m.Collection)
    
	insertResult, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return nil, err
	}
    
    return insertResult.InsertedID, nil
}

func (m *Dao) FindByID(_id string) (map[string]interface{}, error) {
    doc := make(map[string]interface{})
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    objID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		return nil, err
	}
    
    findOne := collection.FindOne(context.TODO(), bson.M{"_id": objID})
	if err := findOne.Err(); err != nil {
		return nil, err
	}
    
    
    err = findOne.Decode(&doc)
	if err != nil {
		return nil, err
	}
    
    return doc, nil
}

func (m *Dao) FindOneWithFilters(qry map[string]interface{}) (map[string]interface{}, error) {
	doc := make(map[string]interface{})
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    findOne := collection.FindOne(context.TODO(), qry)
	if err := findOne.Err(); err != nil {
		return nil, err
	}
    
    err := findOne.Decode(&doc)
	if err != nil {
		return nil, err
	}
    
    return doc, nil
}

func (m *Dao) FindAll() ([]map[string]interface{}, error) {
	docs := make([]map[string]interface{}, 0)
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    // Declare Context type object for managing multiple API requests
    ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
    
    n := make(map[string]interface{}, 0)
    cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
        defer cursor.Close(ctx)
		return nil, err
	}
    
	for cursor.Next(ctx) {
		err := cursor.Decode(&n)
        
        // If there is a cursor.Decode error
        if err != nil {
            fmt.Println("cursor.Next() error:", err)
            os.Exit(1)
        } else {
            docs = append(docs, n)
        }
	}
    
    /*
    err = cursor.All(ctx, res)
    if err != nil {
        fmt.Println(err.Error())
    }
    */
    
    // Dont forget to close the cursor
    defer cursor.Close(context.TODO())
    
    return docs, nil
}

func (m *Dao) FindAllWithFilters(qry map[string]interface{}, sort map[string]int, limit int, after, before string) ([]map[string]interface{}, error) {
	docs := make([]map[string]interface{}, 0)
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    // Declare Context type object for managing multiple API requests
    ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
    
    // Pagination
    findOptions := options.Find() // build a `findOptions`
    //findOptions.SetSort(map[string]int{"when": -1}) // reverse order by `when`
    if sort != nil {
        findOptions.SetSort(sort)
    }
    ///findOptions.SetSkip(0) // skip whatever you want, like `offset` clause in mysql
    if limit > 0 {
        findOptions.SetLimit(int64(limit)) // like `limit` clause in mysql
    }
    
    if after != "" {
        objID, err := primitive.ObjectIDFromHex(after)
        if err != nil {
            return nil, err
        }
        
        qry["_id"] = bson.M{
                "$gt": objID,
            }
    }
    
    if before != "" {
        objID, err := primitive.ObjectIDFromHex(before)
        if err != nil {
            return nil, err
        }
        
        qry["_id"] = bson.M{
                "$lt": objID,
            }
    }
    
    cursor, err := collection.Find(context.TODO(), qry, findOptions)
	if err != nil {
        defer cursor.Close(ctx)
		return nil, err
	}
    
    for cursor.Next(ctx) {
        n := make(map[string]interface{}, 0)
        cursor.Decode(&n)
        
        // If there is a cursor.Decode error
        if err != nil {
            fmt.Println("cursor.Next() error:", err)
            os.Exit(1)
        } else {
            docs = append(docs, n)
        }
    }
    
    // Dont forget to close the cursor
    defer cursor.Close(context.TODO())
    
    return docs, nil
}


/**
func (m *Dao) FindAllWithFiltersAndLimitSkip(qry, prj *map[string]interface{}, Limit, Skip int) ([]map[string]interface{}, int, error) {
    docs := make([]map[string]interface{}, 0)
    
    // Find the number of games won by Dave
    count, err := db.C(m.Collection).Find(&qry).Select(&prj).Count()
    if err != nil {
        log.Println(err)
        return nil, count, err
    }
    
    if err := db.C(m.Collection).Find(&qry).Select(&prj).Limit(Limit).Skip(Skip).All(&docs); err != nil {
        log.Println(err)
		return nil, count, err
	}
	
    return docs, count, nil
}
*/
/** ****** */


/*func (m *Dao) PipeOne(opr interface{}) (map[string]interface{}, error) {
    docs := make(map[string]interface{})
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    if err := db.C(m.Collection).Pipe(&opr).One(&docs); err != nil {
        if err.Error() != "not found" {
            log.Println(err)
        }
		return nil, err
	}
	
    return docs, nil
}*/

func (m *Dao) Aggregate(pipeline interface{}) ([]map[string]interface{}, error) {
    docs := make([]map[string]interface{}, 0)
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    // Declare Context type object for managing multiple API requests
    ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
    
    cursor, err := collection.Aggregate(context.TODO(), pipeline)
	if err != nil {
        defer cursor.Close(ctx)
		return nil, err
	}
    
    for cursor.Next(ctx) {
        n := make(map[string]interface{}, 0)
        err := cursor.Decode(&n)
        
        if err != nil {
            log.Fatal(err)
        }
        
        docs = append(docs, n)
    }
    
    // Dont forget to close the cursor
    defer cursor.Close(context.TODO())
    
    return docs, nil
}


/*func (m *Dao) Update(contact Contact) (err error) {
	err = db.C(m.Collection).UpdateId(contact.ID, &contact)
	return
}*/

func (m *Dao) Update(_id string, doc interface{}) (count int64, err error) {
	var collection = client.Database(m.Database).Collection(m.Collection)
    
    objID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		return -1, err
	}

	resultUpdate, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{
			"$set": &doc,
		},
	)
    if err != nil {
		return -1, err
	}

	return resultUpdate.ModifiedCount, nil
}

/**
func (m *Dao) UpdateFieldInc(id string, field string, inc int) (err error) {
	///if err:= db.C(m.Collection).Update(bson.ObjectIdHex(id), bson.M{"$inc": bson.M{field: inc}}); err != nil{
	if err:= db.C(m.Collection).Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$inc": bson.M{field: inc}}); err != nil{
        if err.Error() != "not found" {
            log.Println(err)
        }
		return err
	}
    
	return nil
}
*/

func (m *Dao) UpdateWithFilters(qry map[string]interface{}, doc interface{}) (count int64, err error) {
    var collection = client.Database(m.Database).Collection(m.Collection)
    
	resultUpdate, err := collection.UpdateMany(
		context.TODO(),
		qry,
		&doc,
	)
    if err != nil {
		return -1, err
	}

	return resultUpdate.ModifiedCount, nil
}

func (m *Dao) UpdateManyWithFilters(qry map[string]interface{}, doc interface{}) (count int64, err error) {
    var collection = client.Database(m.Database).Collection(m.Collection)
    
	resultUpdate, err := collection.UpdateMany(
		context.TODO(),
		qry,
		bson.M{
			"$set": &doc,
		},
	)
    if err != nil {
		return -1, err
	}

	return resultUpdate.ModifiedCount, nil
}

// Delete an Doc from the collection
func (m *Dao) Delete(_id string) (count int64, err error) {
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    objID, err := primitive.ObjectIDFromHex(_id)
	if err != nil {
		return -1, err
	}

	resultDelete, err := collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		return -1, err
	}

	return resultDelete.DeletedCount, nil
}

// Delete all Docs from the collection
func (m *Dao) DeleteAll(qry map[string]interface{}) (count int64, err error) {
    var collection = client.Database(m.Database).Collection(m.Collection)
    
    result, err := collection.DeleteMany(context.TODO(), qry)
    return result.DeletedCount, err
}


// Stats from the collection
// https://mlog.club/article/3362897
func (m *Dao) Stats() (map[string]interface{}, error) {
    db := client.Database(m.Database)
    
    result := db.RunCommand(context.Background(), bson.M{"collStats": m.Collection})
    
    var document bson.M
    err := result.Decode(&document)
    if err != nil {
        return nil, err
    }
    
    return document, nil
}
