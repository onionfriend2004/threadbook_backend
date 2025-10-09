package auth

import "errors"

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrRedisRead       = errors.New("error read from Redis")
	ErrJSONDecode      = errors.New("error decoding JSON")
	ErrRedisConnect    = errors.New("error connect to Redis")
)
