FROM golang:1.15-alpine AS builder
RUN apk update && apk add git
WORKDIR /go/src/garfunkel
COPY . .
RUN go get -v ./...
RUN go build -v .
FROM alpine:3.12
COPY --from=builder /go/src/garfunkel/garfunkel .
CMD ["/garfunkel"]
