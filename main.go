package main

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sys/unix"
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

func parseEnv() error {
	username = os.Getenv("PUSHUSER")
	password = os.Getenv("PUSHPW")
	forwardsite = os.Getenv("FORWARDSITE")
	host = os.Getenv("HOST")
	if username == "" || password == "" || forwardsite == "" {
		return errors.New("missing environment variables")
	}
	datadir = os.Getenv("DATADIR")
	if datadir == "" {
		datadir = "./data"
	}
	return nil
}

func extractfileending(filename string) string {
	if len(filename) >= 7 && filename[len(filename)-7:] == ".tar.gz" {
		return ".tgz"
	}
	if len(filename) >= 8 && filename[len(filename)-8:] == ".tar.zst" {
		return ".tzst"
	}
	ext := filepath.Ext(filename)
	if ext == "" || ext == "." {
		ext = ".dat"
	}
	return ext
}

func generateName(hash *string, filename string, now time.Time) string {
	timedatednow := now.Format("02-01-2006_15:04:05")
	return (timedatednow + "_" + *hash + extractfileending(filename))
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
	if err != nil || file == nil || file.Filename == "" {
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
	filename := generateName(&hashstring, file.Filename, time.Now())
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

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(authreq())
	r.POST("/", saveupload)
	r.NoRoute(forwardtomain)
	return r
}

func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func main() {
	err := parseEnv()
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(datadir); os.IsNotExist(err) {
		panic(err)
	}
	if !writable(datadir) {
		panic("datadir not writable")
	}
	fmt.Printf("fuck")
	r := setupRouter()
	r.Run(":8080")
}
