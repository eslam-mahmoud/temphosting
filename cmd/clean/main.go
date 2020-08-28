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

	db "github.com/eslam-mahmoud/tempstuff/db/files"
	storage "github.com/eslam-mahmoud/tempstuff/storage/files"
	kitlog "github.com/go-kit/kit/log"
)

// $ go run ./cmd/clean/*.go -dbPath="./dbFiles" -storagePath="./uploads"
func main() {
	// read command line flags
	// return pointer
	var dbPath string
	var storagePath string
	flag.StringVar(&dbPath, "dbPath", "", "DB path")
	flag.StringVar(&storagePath, "storagePath", "", "Storage path")
	// parses from os.Args[1:]. Must be called after all flags are defined and before flags are accessed by the program.
	flag.Parse()

	logger := kitlog.With(kitlog.NewJSONLogger(os.Stderr), "ts", kitlog.DefaultTimestampUTC)
	storageService, err := storage.New(logger, storagePath)
	if err != nil {
		logger.Log("message", "could not init storage service", "error", err)
		return
	}
	dbService, err := db.New(logger, dbPath)
	if err != nil {
		logger.Log("message", "could not init DB service", "error", err)
		return
	}

	logger.Log("message", "all services are operational")

	for {
		// get list of the files
		files, err := ioutil.ReadDir(dbPath)
		if err != nil {
			logger.Log(
				"message", "could not open dir",
				"error", err,
			)
			break
		}

		for _, f := range files {
			id := strings.Trim(f.Name(), ".json")
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
					"message", "failed deleting expired file from DB",
					"error", StorageErr,
					"id", id,
				)
			}
			dbErr := dbService.Delete(context.Background(), id)
			if dbErr != nil {
				logger.Log(
					"message", "failed deleting expired file from storage",
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
