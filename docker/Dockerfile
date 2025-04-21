FROM golang:1.22-alpine3.20

RUN mkdir /mitm-proxy

WORKDIR /mitm-proxy

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV GO111MODULE=on

RUN apk add --no-cache openssl

RUN go build -o bin/main ./cmd/main.go

CMD ["./bin/main"]