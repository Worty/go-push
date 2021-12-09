package main

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var username string
var password string
var forwardsite string
var host string
var datadir string

func forwardtomain(c *gin.Context) {
	c.Redirect(302, forwardsite)
	c.Abort()
}

func parseEnv() {
	username = os.Getenv("PUSHUSER")
	password = os.Getenv("PUSHPW")
	forwardsite = os.Getenv("FORWARDSITE")
	host = os.Getenv("HOST")
	if username == "" || password == "" || forwardsite == "" {
		panic("Missing environment variables")
	}
	datadir = os.Getenv("DATADIR")
	if datadir == "" {
		datadir = "./data"
	}
}

func extractfileending(filename string) string {
	return filename[strings.LastIndex(filename, ".")+1:]
}

func generateName(hash *string, filename string) string {
	timedatednow := time.Now().Format("02-01-2006_15:04:05")
	return (timedatednow + "_" + *hash + "." + extractfileending(filename))
}

func authreq() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.PostForm(username) != "" && subtle.ConstantTimeCompare([]byte(c.PostForm(username)), []byte(password)) == 1 {
			c.Next()
		} else {
			fmt.Println("unauthorized")
			forwardtomain(c)
			return
		}
	}
}

func saveupload(c *gin.Context) {
	c.Request.ParseMultipartForm(1024 * 1024 * 1024) // 1 GB
	file, err := c.FormFile("d")
	if file.Filename == "" || err != nil {
		fmt.Println("no file")
		forwardtomain(c)
		return
	}
	hashstring, err := generateMD5(file)
	if err != nil {
		fmt.Println(err.Error())
		forwardtomain(c)
		return
	}
	filename := generateName(&hashstring, file.Filename)
	err = c.SaveUploadedFile(file, datadir+"/"+filename)
	if err != nil {
		fmt.Println(err.Error())
		forwardtomain(c)
		return
	}
	c.String(200, "%s/%s", host, filename)
}

func generateMD5(in *multipart.FileHeader) (string, error) {
	f, err := in.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := md5.New()
	io.Copy(hash, f)
	return hex.EncodeToString(hash.Sum([]byte(""))), nil
}

func main() {
	parseEnv()
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(authreq())
	r.POST("/", saveupload)
	r.NoRoute(forwardtomain)
	r.Run(":8080")
}
