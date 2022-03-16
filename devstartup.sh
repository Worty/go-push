#!/bin/sh
go install github.com/codegangsta/gin@latest
go mod download

PUSHUSER="user" PUSHPW="pw" HOST="localhost" FORWARDSITE="https://google.de" gin -a 8080 -p 8081