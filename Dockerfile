#docker build -t ghcr.io/worty/go-push .
FROM golang:1.17-alpine as go-build
RUN apk add --no-cache gcc libc-dev tzdata
ADD ./go.mod /build/go.mod
ADD ./go.sum /build/go.sum
WORKDIR /build
RUN go mod download

ADD . /build
WORKDIR /build
RUN go build -ldflags "-linkmode external -s -w -extldflags -static" --trimpath -a -o ./main

FROM scratch
COPY --from=go-build /usr/share/zoneinfo /usr/share/zoneinfo
WORKDIR /app/
COPY --from=go-build /build/main /app/
ENV GIN_MODE=release
EXPOSE 8080
CMD ["./main"]