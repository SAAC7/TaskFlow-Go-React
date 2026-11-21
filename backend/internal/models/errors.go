package models

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Domain errors. Every layer speaks these, so the HTTP layer never needs to
// know about MongoDB errors (like mongo.ErrNoDocuments).
var (
	ErrNotFound  = errors.New("task not found")
	ErrInvalidID = errors.New("invalid taks id")
)

// ValidationError carries the list of invalid fields
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "Validation failed"
}

// ParseID converts the hex string of the URL into an ObjectID.
func ParseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}
	return oid, nil
}
