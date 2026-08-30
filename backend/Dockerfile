# Development Stage
FROM golang:1.25-alpine AS development

WORKDIR /app

# Install build dependencies & curl to fetch the precompiled Air binary
RUN apk add --no-cache gcc musl-dev git curl

# Install precompiled Air binary to avoid Go compiler/linker issues inside the build stage
RUN curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.1

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

# Build Stage for Production
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Production Runtime Stage
FROM alpine:3.19 AS production

WORKDIR /app

RUN apk add --no-cache ca-certificates sqlite-libs

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
