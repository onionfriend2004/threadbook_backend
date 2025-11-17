FROM golang:1.24-alpine

# Меняем на стабильные mirrors
RUN sed -i 's/dl-cdn.alpinelinux.org/dl-4.alpinelinux.org/g' /etc/apk/repositories && \
    apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./cmd/threadbook
EXPOSE 8080
CMD ["./main"]