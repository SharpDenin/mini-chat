FROM golang:1.25rc2

WORKDIR /app

# Копируем исходный код
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем приложение
RUN go build -o main cmd/main.go

# Экспортируем порты
EXPOSE 8083 50053 50054

# Запускаем приложение
CMD ["./main"]