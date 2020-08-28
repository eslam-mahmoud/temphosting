package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/eslam-mahmoud/tempstuff/db"
	kitlog "github.com/go-kit/kit/log"
	redisLib "github.com/go-redis/redis/v8"
)

// Repo implement db.Database interface
// should be construct through New()
type Repo struct {
	logger kitlog.Logger
	client *redisLib.Client
}

// New is constructor for Repo
// set the path
func New(l kitlog.Logger, host string) (*Repo, error) {
	logger := kitlog.With(l, "service", "Redis DB")

	client := redisLib.NewClient(&redisLib.Options{
		Addr:     host,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	logger.Log("message", "init redis DB")
	return &Repo{logger: logger, client: client}, nil

}

// Insert create json file from db.Model and save it
func (r *Repo) Insert(ctx context.Context, file *db.Model) error {
	content, err := json.Marshal(file)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, file.ID, content, file.Expiration.Sub(time.Now())).Err()
}

// Get return db entry
func (r *Repo) Get(ctx context.Context, id string) (file db.Model, err error) {
	var val string
	val, err = r.client.Get(ctx, id).Result()
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(val), &file)
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
	_, err = r.client.Del(ctx, id).Result()
	return err
}
