# Используем официальный образ Go
FROM golang:1.24-alpine

# Устанавливаем git (нужен для go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем ВЕСЬ код (включая ./config/)
COPY . .

# Устанавливаем права на чтение (на всякий случай)
RUN chmod -R a+r ./config

# Порт приложения (если нужен)
EXPOSE 8080

# Запуск через go run
CMD ["go", "run", "./cmd/threadbook/main.go"]