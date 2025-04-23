# Используем официальный образ Go как базовый
FROM golang:1.23-alpine as builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Копируем исходники приложения в рабочую директорию
COPY . .

# Собираем приложение
RUN go build -o main ./cmd/main.go

# Начинаем новую стадию сборки на основе минимального образа
FROM alpine:latest

# Добавляем исполняемый файл из первой стадии в корневую директорию контейнера
COPY --from=builder /app/.env /.env
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/main /main

# Открываем порт 8080
EXPOSE 8080
# Запускаем приложение
CMD ["/main"]
