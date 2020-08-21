package main

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const filesDst = "./uploads"

// out DB
var db = make(map[string]uploadedFile)

// file in DB
type uploadedFile struct {
	fileName   string
	expiration time.Time
}

func main() {
	// seed rand with value
	rand.Seed(time.Now().UnixNano())

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": ":)",
		})
	})
	r.POST("/upload", upload)
	r.GET("/file/:id", download)

	// delete expired files
	// every 10 min loop on all files
	// if any of them expired remove it from DB
	go func() {
		for {
			time.Sleep(time.Minute * 10)
			for k, v := range db {
				if v.expiration.Before(time.Now()) {
					os.Remove(path.Join(filesDst, k))
					delete(db, k)
				}
			}
		}
	}()

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// func to handle API req on /:id
func download(c *gin.Context) {
	// get id from req
	id := c.Params.ByName("id")
	// get file from DB
	file, ok := db[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}

	// check on expiration
	if file.expiration.Before(time.Now()) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file expired",
		})

		// if file expired remove it
		os.Remove(path.Join(filesDst, id))
		delete(db, id)
		return
	}

	c.FileAttachment(path.Join(filesDst, id), file.fileName)
}

// func for /upload API end point
func upload(c *gin.Context) {
	// get duration from form
	// validate and set default value if not valid
	duration := c.DefaultPostForm("duration", "10m")
	if duration != "10m" && duration != "1h" {
		duration = "10m"
	}
	fileDuration, err := time.ParseDuration(duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// rename file before saving
	id := generateFilename()
	out, err := os.Create(path.Join(filesDst, id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// save file
	_, err = io.Copy(out, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// save file in DB
	db[id] = uploadedFile{
		fileName:   file.Filename,
		expiration: time.Now().Add(fileDuration),
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "file saved",
		"id":      id,
	})

}

// Get return file
func generateFilename() string {
	nanoTime := time.Now().UnixNano()
	letterRunes := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 16)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return strconv.FormatInt(nanoTime, 10) + "-" + string(b)
}
