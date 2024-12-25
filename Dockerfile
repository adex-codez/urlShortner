FROM golang:1.23-alpine AS build
RUN apk add --no-cache curl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/a-h/templ/cmd/templ@latest && \
    templ generate && \
    curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss && \
    chmod +x tailwindcss && \
    ./tailwindcss -i cmd/web/assets/css/input.css -o cmd/web/assets/css/output.css

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod

# Install necessary runtime libraries for CGO
RUN apk add --no-cache libc6-compat
FROM alpine:3.20.1 AS prod
WORKDIR /app
COPY --from=build /app/main /app/main
EXPOSE ${PORT}
CMD ["./main"]
