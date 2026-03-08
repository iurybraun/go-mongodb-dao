/*
 * Copyright © 2016-2026 Iury Braun
 * Copyright © 2017-2026 VOLKER
 *
 * References:
 *  https://github.com/ReLaboratory/mulbitchorong-backend/blob/ec0bf74c43db49af18540558ec2033a7520b02d9/handle/upload.go
 *  https://github.com/c-jimin/codetech/blob/635798468c09fecb3baa86ba037d8a537c0ad2a7/DaoAndModel/file.go
 */

package go_mongodb_dao

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/iurybraun/go-mongodb-dao/utils/encoder"
)

const defaultFilesBucket = "fs"

// filesBucketName returns the GridFS bucket name, defaulting to "fs".
func (m *Dao) filesBucketName() string {
	if m.Collection != "" {
		return m.Collection
	}
	return defaultFilesBucket
}

// newBucket creates a GridFS bucket for the DAO's database and bucket name.
func (m *Dao) newBucket() (*gridfs.Bucket, error) {
	db := client.Database(m.Database)
	bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName(m.filesBucketName()))
	if err != nil {
		return nil, fmt.Errorf("newBucket: %w", err)
	}
	return bucket, nil
}

// CreateGridFSObject uploads a multipart file to GridFS and returns its
// ObjectID. contentType and path are stored as metadata on the file entry.
func (m *Dao) CreateGridFSObject(
	file multipart.File,
	filename, contentType, path, user string,
) (interface{}, error) {
	bucket, err := m.newBucket()
	if err != nil {
		return nil, errors.Wrap(err, "upload fail")
	}

	md5 := encoder.MD5(file)

	uploadOpts := options.GridFSUpload().SetMetadata(bson.M{
		"path":       path,
		"__v":        0,
		"user":       user,
		"group":      nil,
		"permission": nil,
	})

	uploadStream, err := bucket.OpenUploadStream(filename, uploadOpts)
	if err != nil {
		return nil, fmt.Errorf("CreateGridFSObject — OpenUploadStream: %w", err)
	}
	defer uploadStream.Close()

	if _, err = file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("CreateGridFSObject — file.Seek: %w", err)
	}

	if err = WriteToGridFile(file, uploadStream); err != nil {
		return nil, err
	}

	fileID := uploadStream.FileID

	// Persist md5 and contentType fields that GridFS doesn't set natively.
	filesCollection := client.
		Database(m.Database).
		Collection(m.filesBucketName() + ".files")

	_, err = filesCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": fileID},
		bson.M{"$set": bson.M{"md5": md5, "contentType": contentType}},
	)
	if err != nil {
		return nil, fmt.Errorf("CreateGridFSObject — update md5/contentType: %w", err)
	}

	return fileID, nil
}

// WriteToGridFile copies the contents of file into the GridFS upload stream.
func WriteToGridFile(file multipart.File, uploadStream *gridfs.UploadStream) error {
	defer file.Close()

	reader := bufio.NewReader(file)
	buf := make([]byte, 1024)

	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return errors.New("WriteToGridFile: could not read the input file")
		}
		if n == 0 {
			break
		}
		if _, err := uploadStream.Write(buf[:n]); err != nil {
			return errors.New("WriteToGridFile: could not write to GridFS")
		}
	}

	return nil
}

// ReadGridFSObject downloads a file from GridFS by its ObjectID hex string
// and returns the raw bytes in a buffer.
func (m *Dao) ReadGridFSObject(fileID string) (*bytes.Buffer, error) {
	bucket, err := m.newBucket()
	if err != nil {
		return nil, errors.Wrap(err, "ReadGridFSObject — bucket init fail")
	}

	_id, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return nil, fmt.Errorf("ReadGridFSObject — invalid ObjectID %q: %w", fileID, err)
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if _, err = bucket.DownloadToStream(_id, writer); err != nil {
		return nil, errors.Wrap(err, "ReadGridFSObject — DownloadToStream fail")
	}

	if err = writer.Flush(); err != nil {
		return nil, fmt.Errorf("ReadGridFSObject — flush: %w", err)
	}

	return &buffer, nil
}

// RemoveGridFSObject deletes a file from GridFS by its ObjectID.
func (m *Dao) RemoveGridFSObject(fileID interface{}) error {
	bucket, err := m.newBucket()
	if err != nil {
		return errors.Wrap(err, "RemoveGridFSObject — bucket init fail")
	}

	if err = bucket.Delete(fileID); err != nil {
		return errors.Wrap(err, "RemoveGridFSObject — Delete fail")
	}

	return nil
}
