# go-mongodb-dao

A lightweight Data Access Object (DAO) library for MongoDB in Go, built on top of the official [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver).

## Installation

```bash
go get github.com/iurybraun/go-mongodb-dao
```

## Setup

### Connecting to MongoDB

Call `Connect` once at application startup. The URI can be passed directly or left empty to read from the `MONGO_URL` environment variable (falls back to `mongodb://localhost:27017`).

```go
import dao "github.com/iurybraun/go-mongodb-dao"

func main() {
    // Option 1: explicit URI
    if err := dao.Connect("mongodb://localhost:27017"); err != nil {
        log.Fatal(err)
    }
    defer dao.Disconnect()

    // Option 2: read from MONGO_URL env var
    if err := dao.Connect(""); err != nil {
        log.Fatal(err)
    }
    defer dao.Disconnect()

    // Option 3: URI from config file
    if err := dao.Connect(cfg_ini.LoadKey_string("database", "url")); err != nil {
        log.Fatal(err)
    }
    defer dao.Disconnect()
}
```

### Configuring the Default Database

#### Single-tenant (database fixed at startup)

```go
func main() {
    if err := dao.Connect(""); err != nil {
        log.Fatal(err)
    }
    defer dao.Disconnect()

    dao.SetDatabase(cfg_ini.LoadKey_string("database", "dbname"))
}
```

With the database set globally, omit the `Database` field when creating a `Dao`:

```go
collection := dao.Dao{Collection: "users"}
```

#### Multi-tenant (database per request)

Do not call `SetDatabase`. Instead, pass the database name on each request, typically extracted from the JWT token:

```go
// middleware: store dbName in request context
ctx := context.WithValue(r.Context(), "dbName", claims["dbName"])

// repository: read from context
collection := dao.Dao{
    Database:   db.GetDatabase(r),
    Collection: "users",
}
```

#### Mixed (global default + per-request override)

Both approaches work together. When `Dao.Database` is set, it takes priority over the global value. When empty, the global `dbName` is used as fallback.

```go
dao.SetDatabase("default_db") // global fallback

// uses "default_db"
col1 := dao.Dao{Collection: "logs"}

// overrides to "client_db"
col2 := dao.Dao{Database: "client_db", Collection: "logs"}
```

---

## Usage

### Dao struct

```go
type Dao struct {
    Database   string // optional if SetDatabase() was called
    Collection string
}
```

### Insert

```go
collection := dao.Dao{Collection: "users"}

id, err := collection.Insert(bson.M{
    "name":  "João",
    "email": "joao@email.com",
})
```

### InsertMany

```go
docs := []interface{}{
    bson.M{"name": "João"},
    bson.M{"name": "Maria"},
}

ids, err := collection.InsertMany(docs)
```

### FindByID

```go
doc, err := collection.FindByID("64b1f1a2c3d4e5f6a7b8c9d0")
```

### FindOneWithFilters

```go
doc, err := collection.FindOneWithFilters(map[string]interface{}{
    "email": "joao@email.com",
})
```

### FindAll

```go
docs, err := collection.FindAll()
```

### FindAllWithFilters

```go
docs, err := collection.FindAllWithFilters(
    map[string]interface{}{"active": true}, // query
    map[string]int{"created_at": -1},       // sort
    20,                                      // limit (0 = no limit)
    "",                                      // after  (cursor pagination)
    "",                                      // before (cursor pagination)
)
```

#### Cursor-based pagination

```go
// next page: documents after the last seen ID
docs, err := collection.FindAllWithFilters(filter, sort, 20, lastSeenID, "")

// previous page: documents before the first seen ID
docs, err := collection.FindAllWithFilters(filter, sort, 20, "", firstSeenID)
```

### Aggregate

```go
pipeline := bson.A{
    bson.M{"$match": bson.M{"active": true}},
    bson.M{"$group": bson.M{"_id": "$category", "total": bson.M{"$sum": 1}}},
}

docs, err := collection.Aggregate(pipeline)
```

### Update

Updates a single document by ID using `$set`.

```go
count, err := collection.Update("64b1f1a2c3d4e5f6a7b8c9d0", bson.M{
    "name": "João Silva",
})
```

### UpdateWithFilters

Applies a full update operator document (caller must provide `$set`, `$unset`, etc.).

```go
count, err := collection.UpdateWithFilters(
    map[string]interface{}{"active": false},
    bson.M{"$set": bson.M{"archived": true}},
)
```

### UpdateManyWithFilters

Applies `$set` automatically to all matching documents.

```go
count, err := collection.UpdateManyWithFilters(
    map[string]interface{}{"role": "user"},
    bson.M{"verified": true},
)
```

### Upsert

Updates the first matching document or inserts a new one if no match is found.

```go
count, err := collection.Upsert(
    map[string]interface{}{"email": "joao@email.com"},
    bson.M{"name": "João", "email": "joao@email.com"},
)
```

### Delete

```go
count, err := collection.Delete("64b1f1a2c3d4e5f6a7b8c9d0")
```

### DeleteAll

```go
count, err := collection.DeleteAll(map[string]interface{}{
    "archived": true,
})
```

### Stats

Returns collection-level statistics.

```go
stats, err := collection.Stats()
```

---

## GridFS

GridFS operations use the same `Dao` struct. The `Collection` field sets the bucket name (defaults to `fs`).

### Upload

```go
bucket := dao.Dao{Collection: "uploads"}

fileID, err := bucket.CreateGridFSObject(
    file,         // multipart.File
    "photo.jpg",  // filename
    "image/jpeg", // content type
    "/avatars",   // path metadata
    "user_123",   // user metadata
)
```

### Download

```go
buffer, err := bucket.ReadGridFSObject("64b1f1a2c3d4e5f6a7b8c9d0")
```

### Delete

```go
err := bucket.RemoveGridFSObject(fileID)
```

---

## Helper Functions

```go
// generate a new ObjectID
id := dao.NewObjectID()

// validate an ObjectID string
ok := dao.IsValidObjectId("64b1f1a2c3d4e5f6a7b8c9d0") // true

// get the globally configured database name
dbName := dao.GetDatabase()
```

---

## Environment Variables

| Variable    | Description                              | Default                    |
|-------------|------------------------------------------|----------------------------|
| `MONGO_URL` | MongoDB connection URI used when `Connect("")` is called | `mongodb://localhost:27017` |

---

## License

Copyright © 2016–2026 Iury Braun  
Copyright © 2017–2026 BRAUN TECH
