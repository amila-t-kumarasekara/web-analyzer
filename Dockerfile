ARG GO_VERSION=1.27

FROM golang:${GO_VERSION}-alpine AS deps
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download

FROM deps AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=builder /out/server ./server
USER app
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["./server"]