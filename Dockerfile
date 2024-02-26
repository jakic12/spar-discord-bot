# syntax=docker/dockerfile:1

FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY scraping ./scraping
RUN CGO_ENABLED=0 GOOS=linux go build -o /discord-bot
CMD ["/discord-bot"]