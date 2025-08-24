FROM golang:1.23 AS build
WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" \
    -o /out/app ./cmd

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=build /out/app /app/app

USER nonroot:nonroot
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/app/app"]