#docker build -t ghcr.io/worty/go-push .
FROM golang:1-alpine AS go-build
ENV CGO_ENABLED=0
ENV GOAMD64="v2"
RUN apk add --no-cache tzdata git
ADD ./go.mod /build/go.mod
ADD ./go.sum /build/go.sum
WORKDIR /build
RUN go mod download

ADD . /build
WORKDIR /build
RUN mkdir -p ./data && go env && go test -v && rm -rf ./data
RUN go build -ldflags "-s -w" --trimpath -a -o ./main

FROM scratch
COPY --from=go-build /usr/share/zoneinfo /usr/share/zoneinfo
WORKDIR /app/
COPY --from=go-build /build/main /app/
ENV GIN_MODE=release
EXPOSE 8080
HEALTHCHECK --interval=60s --timeout=15s --start-period=5s --retries=3 CMD [ "/app/main", "--healthcheck" ]
CMD ["./main"]
