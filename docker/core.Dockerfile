FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY apps/core/go.mod apps/core/go.sum ./
RUN go mod download
COPY apps/core/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /velora ./cmd/velora
FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S velora && adduser -S -G velora velora
COPY --from=build /velora /usr/local/bin/velora
USER velora
EXPOSE 8080
ENTRYPOINT ["velora"]
