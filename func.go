/*
 * Copyright © 2016-2026 Iury Braun
 * Copyright © 2017-2026 VOLKER
 */

package go_mongodb_dao

import (
	"regexp"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var objectIDRegex = regexp.MustCompile("^[0-9a-fA-F]{24}$")

// IsValidObjectId reports whether the given string is a valid 24-character
// hexadecimal MongoDB ObjectID.
func IsValidObjectId(objectID string) bool {
	return objectIDRegex.MatchString(objectID)
}

// NewObjectID generates a new unique MongoDB ObjectID.
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}
