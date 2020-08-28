package files

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	kitlog "github.com/go-kit/kit/log"
)

// Storage interface
type Storage interface {
	Save(c context.Context, f *File) error
	Delete(c context.Context, ID string) error
	Get(c context.Context, ID string) (io.Reader, error)
}

// Strg implement Storage interface
type Strg struct {
	logger   kitlog.Logger
	FilesDst string
}

// File represent storage object
type File struct {
	Content    io.Reader
	ID         string
	Expiration time.Time
}

// Delete store file
func (s Strg) Delete(ctx context.Context, id string) error {
	splitedID := strings.Split(id, "-")
	if len(splitedID) != 2 {
		s.logger.Log(
			"message", "invalid id",
			"id", id,
		)
		return errors.New("invalid id")
	}
	unixNano, err := strconv.ParseInt(splitedID[0], 10, 64)
	if err != nil {
		s.logger.Log(
			"message", "failed ParseInt id",
			"error", err,
			"id", id,
		)
		return err
	}
	if time.Now().Before(time.Unix(0, unixNano)) {
		return errors.New("file did not expire yet")
	}

	return os.Remove(filepath.Join(s.FilesDst, id))
}

// Save store file
func (s Strg) Save(ctx context.Context, f *File) error {
	// generate file name on our server
	f.ID = s.generateFilename(f.Expiration)

	// create output file
	out, err := os.Create(filepath.Join(s.FilesDst, f.ID))
	if err != nil {
		return err
	}
	defer out.Close()

	// save the source in the output file
	_, err = io.Copy(out, f.Content)
	// return err which will be nil if copy did not cause issues
	return err
}

// Get return file
func (s Strg) Get(ctx context.Context, ID string) (io.Reader, error) {
	return os.Open(filepath.Join(s.FilesDst, ID))
}

// Get return file
func (s Strg) generateFilename(expiration time.Time) string {
	letterRunes := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return strconv.FormatInt(expiration.UnixNano(), 10) + "-" + string(b)
}

// New constructor for Strg
func New(l kitlog.Logger, path string) (*Strg, error) {
	logger := kitlog.With(l, "service", "Storage")

	if path == "" {
		l.Log("message", "Could not init service", "error", "empty path")
		return nil, errors.New("Storage path is needed")
	}
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		l.Log("message", "Could not init service and MkdirAll", "error", err)
		return nil, err
	}
	return &Strg{
		logger:   logger,
		FilesDst: path,
	}, nil
}
