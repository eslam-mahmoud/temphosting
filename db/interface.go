package db

import (
	"context"
	"time"
)

// Database interface
type Database interface {
	Insert(ctx context.Context, file *Model) error
	Get(ctx context.Context, id string) (file Model, err error)
	Delete(ctx context.Context, id string) error
}

// Model struct represent DB record of file
type Model struct {
	ID         string    `json:"id"`
	FileName   string    `json:"file_name"`
	Expiration time.Time `json:"expiration"`
	Length     int64     `json:"length"`
}
