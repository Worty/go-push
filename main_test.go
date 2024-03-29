package main

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var testfile = "testfile.bin"

func TestFolderCreation(t *testing.T) {
	path := "./data"
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			t.Error(err)
		}
	}
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	file.Close()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, fi.Name())
	if err != nil {
		return nil, err
	}
	part.Write(fileContents)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", uri, body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return r, err
}

func TestHealth(t *testing.T) {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/healthcheck", nil)
	if err != nil {
		t.Error("error creating request:", err)
	}
	router.ServeHTTP(w, req)
	if w.Code != 200 && w.Body.String() != "OK" {
		t.Errorf("Health failed: %d, Body: %v", w.Code, w.Body.String())
	}
}

func TestPush(t *testing.T) {
	username := "test"
	password := "test"
	t.Setenv("PUSHUSER", username)
	t.Setenv("PUSHPW", password)
	forward := "http://localhost:8080"
	t.Setenv("FORWARDSITE", forward)
	t.Setenv("DATADIR", "./data")
	parseEnv()

	router := setupRouter()

	w := httptest.NewRecorder()
	extraParams := map[string]string{
		username: password,
	}
	req, err := newfileUploadRequest("/", extraParams, "d", testfile)
	if err != nil {
		t.Error("error creating request:", err)
	}
	router.ServeHTTP(w, req)
	if w.Code == 200 {
		body := w.Body.String()
		wanthash := getSystemHashForFile(testfile)
		if !strings.Contains(body, wanthash) {
			t.Errorf("hash not found in response got %v want %v\n", body, wanthash)
		}
		return
	}
	t.Errorf("Push failed: %d, Body: %v", w.Code, w.Body.String())
}

func TestPushWithWrongCreds(t *testing.T) {
	username := "test"
	password := "test"
	t.Setenv("PUSHUSER", username)
	t.Setenv("PUSHPW", password)
	forward := "http://localhost:8080"
	t.Setenv("FORWARDSITE", forward)
	t.Setenv("DATADIR", "./data")
	parseEnv()

	router := setupRouter()

	w := httptest.NewRecorder()
	extraParams := map[string]string{
		username: "failpw",
	}
	req, err := newfileUploadRequest("/", extraParams, "d", testfile)
	if err != nil {
		t.Error("error creating request:", err)
	}
	router.ServeHTTP(w, req)
	if w.Code == 302 && w.Result().Header["Location"][0] == forward {
		return
	}
	t.Errorf("Push failed: %d, Body: %v", w.Code, w.Body.String())
}

func TestEmptyEnv(t *testing.T) {
	err := parseEnv()
	if err != nil && err.Error() == "missing environment variables" {
		return
	}
	t.Error("parseEnv failed:", err)
}

func TestExtractfileending_txt(t *testing.T) {
	want := ".txt"
	got := extractfileending("test.txt")
	if got != want {
		t.Error("extractfileending() failed: got", got, "want", want)
	}
}

func TestExtractfileending_tar_gz(t *testing.T) {
	want := ".tgz"
	got := extractfileending("test.tar.gz")
	if got != want {
		t.Error("extractfileending() failed: got", got, "want", want)
	}
}
func TestExtractfileending_zstd(t *testing.T) {
	want := ".tzst"
	got := extractfileending("test.tar.zst")
	if got != want {
		t.Error("extractfileending() failed: got", got, "want", want)
	}
}

func TestExtractfileending_dotdottxt(t *testing.T) {
	want := ".txt"
	got := extractfileending("test..txt")
	if got != want {
		t.Error("extractfileending() failed: got", got, "want", want)
	}
}
func TestExtractfileending_dotdotdot(t *testing.T) {
	want := ".dat"
	got := extractfileending("bla...")
	if got != want {
		t.Error("extractfileending() failed: got", got, "want", want)
	}
}
func TestExtractfileending_noext(t *testing.T) {
	want := ".dat"
	got := extractfileending("bla")
	if got != want {
		t.Error("extractfileending() failed: got", got, "want", want)
	}
}

func TestGenerateName(t *testing.T) {
	hash := "dfde499e6b44b8757c4b58c3c3768236"
	filename := "test.txt"
	now, _ := time.Parse("02-01-2006_15:04:05", "02-01-2020_14:31:05")
	want := "02-01-2020_14:31:05_dfde499e6b44b8757c4b58c3c3768236.txt"
	got := generateName(hash, filename, now)
	if got != want {
		t.Error("generateName() failed: got", got, "want", want)
	}
}

func getSystemHashForFile(filename string) string {
	cmd := exec.Command("md5sum", filename)
	res, err := cmd.Output()
	if err != nil {
		return ""
	}
	restr := string(res)
	return restr[:strings.Index(restr, " ")]
}
