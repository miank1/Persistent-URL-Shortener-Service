FROM golang:1.25.3-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /url-shortener .

FROM alpine:latest

WORKDIR /app

ENV PORT=8080
ENV DB_PATH=/app/data/urls.db

RUN mkdir -p /app/data

COPY --from=build /url-shortener /url-shortener

EXPOSE 8080

CMD ["/url-shortener"]
