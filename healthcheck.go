package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func healthcheck() {
	code := 0
	resp, err := http.Get("http://127.0.0.1:8080/healthcheck")
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("Status Code")
		code = 1
	} else {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			if string(body) != "OK" {
				fmt.Printf("Body: %v\n", string(body))
				code = 1
			}
		} else {
			fmt.Printf("Error: %v\n", err.Error())
			code = 1
		}
	}
	os.Exit(code)
}
