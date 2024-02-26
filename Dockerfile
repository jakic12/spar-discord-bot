# syntax=docker/dockerfile:1

FROM golang:1.22

COPY ./go.mod ./go.sum ./.env  ./
RUN go mod download
COPY *.go ./
RUN mkdir -p scraping
RUN mkdir -p data
COPY scraping/* ./scraping/
RUN CGO_ENABLED=0 GOOS=linux go build -o /discord-bot
CMD ["/discord-bot"]