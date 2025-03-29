FROM golang:1.22-alpine3.20

RUN mkdir /mitm-proxy

WORKDIR /mitm-proxy

COPY go.mod ./
RUN go mod download

COPY . .

ENV GO111MODULE=on

RUN chmod +x internal/scripts/gen_cert.sh

RUN go build -o bin/main ./app/main.go

CMD ["./bin/main"]