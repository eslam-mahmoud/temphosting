package files

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/eslam-mahmoud/tempstuff/db"
	kitlog "github.com/go-kit/kit/log"
)

// Repo implement db.Database interface
// should be construct through New()
type Repo struct {
	logger kitlog.Logger
	path   string
}

// New is constructor for Repo
// set the path
func New(l kitlog.Logger, path string) (*Repo, error) {
	logger := kitlog.With(l, "Service", "DB")

	if path == "" {
		l.Log("message", "Could not init service", "error", "empty path")
		return nil, errors.New("DB path is needed")
	}
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		l.Log("message", "Could not init service and MkdirAll", "error", err)
		return nil, err
	}
	return &Repo{logger: logger, path: path}, nil
}

// Insert create json file from db.Model and save it
func (r *Repo) Insert(ctx context.Context, file *db.Model) error {
	// convert the struct to json file
	content, err := json.Marshal(file)
	if err != nil {
		r.logger.Log("message", "failed to marshal json", "error", err)
		return err
	}
	// save the file
	return ioutil.WriteFile(filepath.Join(r.path, file.ID+".json"), content, 0644)
}

// Get return db entry
func (r *Repo) Get(ctx context.Context, id string) (file db.Model, err error) {
	// get file content
	bytes, err := ioutil.ReadFile(filepath.Join(r.path, id+".json"))
	if err != nil {
		r.logger.Log(
			"message", "Could not get file",
			"error", err,
			"id", id,
		)
		return file, errors.New("could not get file")
	}

	err = json.Unmarshal(bytes, &file)
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
	return os.Remove(filepath.Join(r.path, id+".json"))
}
