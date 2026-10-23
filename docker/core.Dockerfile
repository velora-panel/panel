FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.work go.work.sum ./
COPY apps/core/go.mod apps/core/go.sum ./apps/core/
COPY apps/cli/go.mod apps/cli/go.sum ./apps/cli/
WORKDIR /src/apps/cli
RUN go mod download
COPY apps/core/ /src/apps/core/
COPY apps/cli/ /src/apps/cli/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /velora ./cmd/velora
FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -S velora && adduser -S -G velora velora
COPY --from=build /velora /usr/local/bin/velora
USER velora
EXPOSE 8080
ENTRYPOINT ["velora"]
CMD ["serve"]
