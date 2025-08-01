package state

import "errors"

var (
	ErrStateNotFound = errors.New("state key not found")

	ErrRedisCmdFailed = errors.New("redis command failed")

	ErrSerializationFailed = errors.New("failed to serialize object")

	ErrDeserializationFailed = errors.New("failed to deserialize object")
)
