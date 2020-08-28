package main

import (
	"os"

	itemService "github.com/eslam-mahmoud/tempstuff"
	db "github.com/eslam-mahmoud/tempstuff/db/files"
	dbMemory "github.com/eslam-mahmoud/tempstuff/db/memory"
	dbRedis "github.com/eslam-mahmoud/tempstuff/db/redis"
	storage "github.com/eslam-mahmoud/tempstuff/storage/files"
	"github.com/gin-gonic/gin"
	kitlog "github.com/go-kit/kit/log"
)

// RedisHost env var name
const RedisHost = "REDIS_HOST"

func main() {
	// HACK TO KEEP THE IMPORT STATEMENT
	_ = dbMemory.Repo{}
	_ = db.Repo{}
	_ = dbRedis.Repo{}

	// init service
	loggerService := kitlog.With(kitlog.NewJSONLogger(os.Stderr), "ts", kitlog.DefaultTimestampUTC)
	storageService, err := storage.New(loggerService, "./uploads")
	if err != nil {
		loggerService.Log("message", "could not init storage service", "error", err)
	}
	// dbService, err := db.New(loggerService, "./dbFiles")
	// if err != nil {
	// 	loggerService.Log("message", "could not init DB service", "error", err)
	// }
	dbService, err := dbRedis.New(loggerService, os.Getenv(RedisHost))
	if err != nil {
		loggerService.Log("message", "could not init redis DB service", "error", err)
	}
	s := itemService.New(loggerService, storageService, dbService)

	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()
	// // LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// // By default gin.DefaultWriter = os.Stdout
	// router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

	// 	// your custom format
	// 	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 			param.ClientIP,
	// 			param.TimeStamp.Format(time.RFC1123),
	// 			param.Method,
	// 			param.Path,
	// 			param.Request.Proto,
	// 			param.StatusCode,
	// 			param.Latency,
	// 			param.Request.UserAgent(),
	// 			param.ErrorMessage,
	// 	)
	// }))
	// https://github.com/gin-gonic/gin/blame/c6d6df6d5ada990c902c51a54b9c4c6f21f87840/README.md#L2056
	// gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
	// 	log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	// }
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
