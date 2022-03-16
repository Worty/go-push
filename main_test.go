package main

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileContents, err := ioutil.ReadAll(file)
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
	req, err := newfileUploadRequest("/", extraParams, "d", "testfile.bin")
	if err != nil {
		t.Error("error creating request:", err)
	}
	router.ServeHTTP(w, req)
	if w.Code == 200 {
		t.Logf("%v\n", w.Body.String())
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
	req, err := newfileUploadRequest("/", extraParams, "d", "testfile.bin")
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

func TestExtractfileending_dotdottxt(t *testing.T) {
	want := ".txta"
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
	got := generateName(&hash, filename, now)
	if got != want {
		t.Error("generateName() failed: got", got, "want", want)
	}
}
