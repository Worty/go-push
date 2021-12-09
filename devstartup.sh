#!/bin/sh
go install github.com/codegangsta/gin@latest
go mod download

PUSHUSER="user" PUSHPW="pw" HOST="localhost" FORWARDSITE="https://worty.de" FORWARDSITE="https://worty.de" gin -a 8080 -p 8081