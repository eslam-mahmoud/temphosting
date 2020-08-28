package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	db "github.com/eslam-mahmoud/tempstuff/db"
	filesDB "github.com/eslam-mahmoud/tempstuff/db/files"
	redisDB "github.com/eslam-mahmoud/tempstuff/db/redis"
	storage "github.com/eslam-mahmoud/tempstuff/storage/files"
	kitlog "github.com/go-kit/kit/log"
)

// $ go run ./cmd/clean/*.go -dbType=files -dbPath="./dbFiles" -storagePath="./uploads"
// $ go run ./cmd/clean/*.go -dbType="redis" -redisHost="redis:6379" -storagePath="/go/src/app/uploads/"
func main() {
	// read command line flags
	// return pointer
	var dbType string
	var dbPath string
	var redisHost string
	var storagePath string
	var dbService db.Database

	flag.StringVar(&dbType, "dbType", "", "DB Type (redis, files)")
	flag.StringVar(&dbPath, "dbPath", "", "DB path for files DB")
	flag.StringVar(&redisHost, "redisHost", "", "DB host url for redis DB")
	flag.StringVar(&storagePath, "storagePath", "", "Storage path")
	// parses from os.Args[1:]. Must be called after all flags are defined and before flags are accessed by the program.
	flag.Parse()

	logger := kitlog.With(kitlog.NewJSONLogger(os.Stderr), "ts", kitlog.DefaultTimestampUTC)
	storageService, err := storage.New(logger, storagePath)
	if err != nil {
		logger.Log("message", "could not init storage service", "error", err)
		return
	}
	if dbType == "files" {
		dbService, err = filesDB.New(logger, dbPath)
		if err != nil {
			logger.Log("message", "could not init files DB service", "error", err)
			return
		}
		logger.Log("message", "Init files DB service", "error", err)
	} else if dbType == "redis" {
		dbService, err = redisDB.New(logger, redisHost)
		if err != nil {
			logger.Log("message", "could not init redis DB service", "error", err)
		}
		logger.Log("message", "Init redis DB service", "error", err)
	} else {
		logger.Log("message", "could not init DB service", "error", err)
		return
	}

	logger.Log("message", "all services are operational")

	for {
		// get list of the files
		files, err := ioutil.ReadDir(storagePath)
		if err != nil {
			logger.Log(
				"message", "could not open dir",
				"error", err,
			)
			break
		}

		for _, f := range files {
			id := f.Name()
			// get expiration date from file name
			splitedID := strings.Split(id, "-")
			if len(splitedID) != 2 {
				logger.Log(
					"message", "invalid id",
					"id", id,
				)
				continue
			}
			unixNano, err := strconv.ParseInt(splitedID[0], 10, 64)
			if err != nil {
				logger.Log(
					"message", "failed ParseInt id",
					"error", err,
					"id", id,
				)
				continue
			}

			// if did not expire continue to next file
			if time.Now().Before(time.Unix(0, unixNano)) {
				logger.Log(
					"message", "Skipping file",
					"id", id,
					"expiration", time.Unix(0, unixNano),
				)
				continue
			}

			logger.Log(
				"message", "File expired",
				"id", id,
				"expiration", time.Unix(0, unixNano),
			)

			// file expired remove it from DB and storage
			StorageErr := storageService.Delete(context.Background(), id)
			if StorageErr != nil {
				logger.Log(
					"message", "failed deleting expired file from storage",
					"error", StorageErr,
					"id", id,
				)
			}
			dbErr := dbService.Delete(context.Background(), id)
			if dbErr != nil {
				logger.Log(
					"message", "failed deleting expired file from DB",
					"error", dbErr,
					"id", id,
				)
			}

			if dbErr == nil && StorageErr == nil {
				logger.Log(
					"message", "File deleted",
					"id", id,
				)
			}
		}

		// loop every 10 min
		logger.Log("message", "loop finished will sleep for 10m")
		time.Sleep(time.Minute * 10)
	}
	fmt.Println("main function will exit now")
}
