package main

import (
	"os"

	itemService "github.com/eslam-mahmoud/tempstuff"
	db "github.com/eslam-mahmoud/tempstuff/db/files"
	dbMemory "github.com/eslam-mahmoud/tempstuff/db/memory"
	storage "github.com/eslam-mahmoud/tempstuff/storage/files"
	"github.com/gin-gonic/gin"
	kitlog "github.com/go-kit/kit/log"
)

func main() {
	// HACK TO KEEP THE IMPORT STATEMENT
	_ = dbMemory.Repo{}
	_ = db.Repo{}

	// init service
	loggerService := kitlog.With(kitlog.NewJSONLogger(os.Stderr), "ts", kitlog.DefaultTimestampUTC)
	storageService, err := storage.New(loggerService, "./uploads")
	if err != nil {
		loggerService.Log("message", "could not init storage service", "error", err)
	}
	dbService, err := db.New(loggerService, "./dbFiles")
	if err != nil {
		loggerService.Log("message", "could not init DB service", "error", err)
	}
	// dbService, err := dbMemory.New(loggerService)
	// if err != nil {
	// 	loggerService.Log("message", "could not init DB service", "error", err)
	// }
	s := itemService.New(loggerService, storageService, dbService)

	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()
	r.Use(setItemService(s))

	// setup routs
	r.GET("/ping", pong)
	r.GET("/i/:id", getItem)
	r.POST("/upload", upload)

	// TODO read from env
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func setItemService(s *itemService.Srvs) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("itemService", s)
		// before request
		c.Next()
		// after request
	}
}
