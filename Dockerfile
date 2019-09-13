FROM golang:1.12-alpine
RUN apk update && apk add git
WORKDIR /go/src/garfunkel
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
CMD ["garfunkel"]
