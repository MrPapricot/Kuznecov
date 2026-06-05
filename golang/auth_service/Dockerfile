FROM golang:1.26.2 AS builder

WORKDIR /usr/src/app

COPY ./go.work ./go.work

COPY ./shared/auth ./shared/auth
COPY ./shared/go.mod ./shared/go.mod

COPY ./utils/go.mod ./utils/go.mod
COPY ./utils/env_parser.go ./utils/env_parser.go

COPY ./auth_service/go.mod ./auth_service/go.mod
COPY ./auth_service/src ./auth_service/src

COPY ./db_adapter/go.mod ./db_adapter/go.mod
COPY ./db_adapter/src ./db_adapter/src

COPY ./backend/go.mod ./backend/go.mod
COPY ./backend/src ./backend/src

WORKDIR /usr/src/app/auth_service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./target/auth_service ./src

FROM alpine:3.17 AS start

COPY --from=builder /usr/src/app/auth_service/target/auth_service /usr/local/bin/auth_service

CMD ["auth_service"]
