# syntax=docker/dockerfile:1

FROM golang:1.22.12-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/scaffold-api .

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
	&& addgroup -S app \
	&& adduser -S -G app app

COPY --from=builder /out/scaffold-api /usr/local/bin/scaffold-api
COPY configs/config.yaml.example /app/configs/config.yaml.example

USER app

EXPOSE 8080

ENTRYPOINT ["scaffold-api"]
CMD ["serve"]
