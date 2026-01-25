# Stage 1: Build
FROM golang:1.23-alpine AS build
LABEL "language"="docker"
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
# Add -ldflags="-s -w" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./main.go

# Stage 2: Final
FROM alpine:latest
# Install necessary certificates and timezone data for HTTPS and correct time
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /app/main .

# Expose port (good practice)
EXPOSE 8080

CMD ["./main"]  