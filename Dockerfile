FROM golang:1.12-stretch AS builder

WORKDIR /go/src/github.com/glitchd/glitchd-server

COPY . .

RUN GO111MODULE=on go get ./...
RUN GO111MODULE=on go generate ./...
RUN GO111MODULE=on GOOS=linux GOARCH=386 go build -o gql-server server/server.go


FROM golang:1.12-alpine

WORKDIR /app/
COPY --from=builder /go/src/github.com/glitchd/gqlgen-starter .

CMD /app/gql-server
