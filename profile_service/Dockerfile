FROM golang:1.25rc2 AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY profile_service/go.mod profile_service/go.sum ./
COPY shared-proto/go.mod shared-proto/go.sum ./shared-proto/

RUN go mod edit -replace shared-proto=./shared-proto
RUN go mod download

COPY profile_service/ ./
COPY shared-proto/ ./shared-proto/

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --output ./docs --parseDependency --parseInternal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
EXPOSE 8083 50053 50054
CMD ["./main"]