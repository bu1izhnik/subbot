# Dockerfile for goose migrations

FROM golang:1.22-alpine AS build
RUN go install github.com/pressly/goose/cmd/goose@latest

FROM alpine:3.20
RUN apk add --no-cache postgresql-client
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
ENTRYPOINT ["/bin/sh","-c"]
CMD ["goose -dir=/migrations status"]