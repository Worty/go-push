package main

import (
	"fmt"
	"io"
	"net/http"
)

func healthcheck() error {
	resp, err := http.Get("http://127.0.0.1:8080/healthcheck")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code not ok: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(body) != "OK" {
		return fmt.Errorf("body not ok: %s", string(body))
	}
	return nil
}
