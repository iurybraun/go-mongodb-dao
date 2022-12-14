/*
 * Copyright © 2016-2022 Iury Braun
 * Copyright © 2017-2022 Neo7i
 * 
 * Alt:  Id bson.ObjectId  ==>  id interface{}
 * 
 * https://github.com/ReLaboratory/mulbitchorong-backend/blob/ec0bf74c43db49af18540558ec2033a7520b02d9/handle/upload.go
 * https://github.com/c-jimin/codetech/blob/635798468c09fecb3baa86ba037d8a537c0ad2a7/DaoAndModel/file.go
 */

package go_mongodb_dao

import (
    "log"
    //"time"
    "io"
    "bufio"
	"bytes"
    //"net/http"
	"context"
	"mime/multipart"
    //"path/filepath"
    
    "github.com/pkg/errors"
    
    //"fmt"
    //"image"
    //"image/draw"
    //"strconv"
    //"image/jpeg"
    //"image/color"
    //"encoding/base64"
    //"os"
    //"io"
    //"log"
    //"bytes"
    //"net/http"
    //"github.com/nfnt/resize"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
    
	"github.com/iurybraun/go-mongodb-dao/utils/encoder"
)

const defaultFilesBucket = "fs"

func (m *Dao) CreateGridFSObject(file multipart.File, filename, contentType, path, user string) (interface{}, error) {
    //var conn db.Connection
    var collection = client.Database(m.Database)  //.Collection(m.Collection)
	//bucket, err := GetBucket(fileBucket, &conn)
    var filesBucket string
    if m.Collection != "" {
        filesBucket = m.Collection
    } else {
        filesBucket = defaultFilesBucket
    }
    bucket, err := gridfs.NewBucket(collection, options.GridFSBucket().SetName(filesBucket))
	//defer conn.Client.Disconnect(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "upload fail")
	}
	md5 := encoder.MD5(file)
    log.Println(md5)
    //"md5": md5, "contentType": contentType, 
    opts := options.GridFSUpload().SetMetadata(bson.M{"path" : path, "__v" : 0, "user" : user, "group" : nil, "permission" : nil})
    uploadStream, uploadStreamErr := bucket.OpenUploadStream(filename, opts)
	if uploadStreamErr != nil {
		log.Println(uploadStreamErr)
        return nil, uploadStreamErr
	}
    file.Seek(0, 0)
    defer uploadStream.Close()
    writeErr := WriteToGridFile(file, uploadStream)
    if writeErr != nil {
        return nil, writeErr
    }
    
    
    fileID := uploadStream.FileID
    
    /* update md5/contentType */
    var collectionUpdate = client.Database(m.Database).Collection(filesBucket + ".files")
    _, errUpdate := collectionUpdate.UpdateOne(
		context.Background(),
		bson.M{"_id": fileID},
		bson.M{
			"$set": bson.M{"md5" : md5, "contentType" : contentType},
		},
	)
    if errUpdate != nil {
		return nil, errUpdate
	}
    /* end update */
    
    return fileID, nil
}

func WriteToGridFile(file multipart.File, uploadStream *gridfs.UploadStream) error {
	reader := bufio.NewReader(file)
	defer func() { file.Close() }()
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return errors.New("Could not read the input file")
		}
		if n == 0 {
			break
		}
		if _, err := uploadStream.Write(buf[:n]); err != nil {
			return errors.New("Could not write to GridFs")
		}
	}
	uploadStream.Close()
	return nil
}

func (m *Dao) ReadGridFSObject(fileID string) (*bytes.Buffer, error) {
    //var conn db.Connection
	var collection = client.Database(m.Database)  //.Collection(m.Collection)
    //bucket, err := GetBucket(fileBucket, &conn)
    var filesBucket string
    if m.Collection != "" {
        filesBucket = m.Collection
    } else {
        filesBucket = defaultFilesBucket
    }
    bucket, err := gridfs.NewBucket(collection, options.GridFSBucket().SetName(filesBucket))
    //defer conn.Client.Disconnect(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "getFile fail")
	}
    
	var buffer bytes.Buffer
	w := bufio.NewWriter(&buffer)
	_id, _ := primitive.ObjectIDFromHex(fileID)
    
	if _, err = bucket.DownloadToStream(_id, w); err != nil {
		return nil, errors.Wrap(err, "getFile - bucket.DownloadToStream fail")
	}
    
	return &buffer, nil
}

func (m *Dao) RemoveGridFSObject(fileID interface{}) error {
    var collection = client.Database(m.Database)  //.Collection(m.Collection)
    var filesBucket string
    if m.Collection != "" {
        filesBucket = m.Collection
    } else {
        filesBucket = defaultFilesBucket
    }
    bucket, err := gridfs.NewBucket(collection, options.GridFSBucket().SetName(filesBucket))
    //defer conn.Client.Disconnect(context.Background())
	if err != nil {
		return errors.Wrap(err, "getFile fail")
	}
    
    if err = bucket.Delete(fileID); err != nil {
		return errors.Wrap(err, "getFile - bucket.DownloadToStream fail")
	}
    
    return nil
}
