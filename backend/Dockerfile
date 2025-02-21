FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV GIN_MODE=release

RUN go build -o main .

EXPOSE 8080

CMD ["./main"] 