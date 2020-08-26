package memory

import (
	"context"
	"errors"
	"time"

	"github.com/eslam-mahmoud/tempstuff/db"
	kitlog "github.com/go-kit/kit/log"
)

// Repo implement db.Database interface
// should be construct through New()
type Repo struct {
	logger kitlog.Logger
	db     map[string]db.Model
}

// New is constructor for Repo
// set the path
func New(l kitlog.Logger) (*Repo, error) {
	logger := kitlog.With(l, "Service", "In memeory DB")

	return &Repo{logger: logger, db: make(map[string]db.Model)}, nil
}

// Insert create json file from db.Model and save it
func (r *Repo) Insert(ctx context.Context, file *db.Model) error {
	r.db[file.ID] = *file
	return nil
}

// Get return db entry
func (r *Repo) Get(ctx context.Context, id string) (file db.Model, err error) {
	// get file from DB
	file, ok := r.db[id]
	if !ok {
		err = errors.New("file not found")
	}
	return
}

// Delete remove entry from DB
func (r *Repo) Delete(ctx context.Context, id string) error {
	file, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	if time.Now().Before(file.Expiration) {
		return errors.New("file did not expire yet")
	}
	delete(r.db, id)
	return nil
}
