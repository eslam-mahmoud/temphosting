package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	itemService "github.com/eslam-mahmoud/tempstuff"
	"github.com/gin-gonic/gin"
)

func getItem(c *gin.Context) {
	srvc := c.MustGet("itemService").(*itemService.Srvs)
	id := c.Params.ByName("id")
	file, err := srvc.Get(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	} else {
		extraHeaders := map[string]string{
			"Content-Disposition": `attachment; filename="` + file.FileName + `"`,
		}

		c.DataFromReader(http.StatusOK, file.Length, "application/octet-stream", file.Body, extraHeaders)
	}
}

func pong(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func home(c *gin.Context) {
	path, err := os.Getwd()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.File(filepath.Join(path, "cmd/webserver/view/index.html"))
}

func upload(c *gin.Context) {
	srvc := c.MustGet("itemService").(*itemService.Srvs)

	// get file duration
	duration := c.DefaultPostForm("duration", "10m")
	if duration != "1s" && duration != "10m" && duration != "1h" && duration != "24h" {
		duration = "10m"
	}
	fileDuration, err := time.ParseDuration(duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// get the file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// max file size 10 MB
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Max file size is 10MB, %vB given", file.Size)})
		return
	}

	// open the file, just open do not read
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileName, err := srvc.Upload(c, itemService.Item{
		Body:       src,
		Expiration: time.Now().Add(fileDuration),
		FileName:   file.Filename,
		Length:     file.Size,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": fileName})
}
