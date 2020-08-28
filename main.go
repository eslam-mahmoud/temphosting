package service

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eslam-mahmoud/tempstuff/db"
	storage "github.com/eslam-mahmoud/tempstuff/storage/files"
	kitlog "github.com/go-kit/kit/log"
)

// Item struct represent the stored items
type Item struct {
	ID         string
	Body       io.Reader
	FileName   string
	Expiration time.Time
	Length     int64
}

// Srvs implement Service interface
type Srvs struct {
	logger  kitlog.Logger
	storage storage.Storage
	db      db.Database
}

// Service is interface define the service functions
type Service interface {
	Get(ctx context.Context, id string) (Item, error)
	Upload(ctx context.Context, opts Item) (string, error)
}

// Get return item by id
func (s Srvs) Get(ctx context.Context, id string) (Item, error) {
	// get expiration from id
	splitedID := strings.Split(id, "-")
	if len(splitedID) != 2 {
		s.logger.Log(
			"message", "invalid id",
			"id", id,
		)
		return Item{}, errors.New("invalid id")
	}
	unixNano, err := strconv.ParseInt(splitedID[0], 10, 64)
	if err != nil {
		s.logger.Log(
			"message", "failed ParseInt id",
			"error", err,
			"id", id,
		)
		return Item{}, err
	}

	// validate expiration
	// if expired delete and ignore errors
	if time.Now().After(time.Unix(0, unixNano)) {
		// err := s.db.Delete(ctx, id)
		// if err != nil {
		// 	s.logger.Log(
		// 		"message", "failed deleting expired file from DB",
		// 		"error", err,
		// 		"id", id,
		// 	)
		// }
		// err = s.storage.Delete(ctx, id)
		// if err != nil {
		// 	s.logger.Log(
		// 		"message", "failed deleting expired file from storage",
		// 		"error", err,
		// 		"id", id,
		// 	)
		// }
		return Item{}, errors.New("File expired")
	}

	// get from DB
	fileModel, err := s.db.Get(ctx, id)
	if err != nil {
		return Item{}, err
	}
	// get from storage
	file, err := s.storage.Get(ctx, id)
	if err != nil {
		return Item{}, err
	}

	return Item{
		ID:         id,
		Body:       file,
		Expiration: fileModel.Expiration,
		FileName:   fileModel.FileName,
		Length:     fileModel.Length,
	}, nil
}

// Upload creates and save item and return it
func (s Srvs) Upload(ctx context.Context, i Item) (string, error) {
	// TODO validate file size

	f := &storage.File{
		Content:    i.Body,
		Expiration: i.Expiration,
	}
	err := s.storage.Save(ctx, f)
	if err != nil {
		return "", err
	}

	err = s.db.Insert(ctx, &db.Model{
		FileName:   i.FileName,
		ID:         f.ID,
		Expiration: f.Expiration,
		Length:     i.Length,
	})
	if err != nil {
		return "", err
	}

	return f.ID, nil
}

// New construct Srvs
func New(logger kitlog.Logger, s storage.Storage, db db.Database) *Srvs {
	// TODO check for null validation on parrams
	return &Srvs{
		logger:  logger,
		storage: s,
		db:      db,
	}
}
