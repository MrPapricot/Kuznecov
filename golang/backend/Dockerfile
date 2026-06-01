FROM golang:1.26.2 AS builder

WORKDIR /usr/src/app

COPY ./go.work ./

COPY ./auth_service/go.mod ./auth_service/go.mod
COPY ./auth_service/src ./auth_service/src

COPY ./db_adapter/go.mod ./db_adapter/go.mod
COPY ./db_adapter/src ./db_adapter/src

COPY ./backend/go.mod ./backend/go.mod
COPY ./backend/src ./backend/src

WORKDIR /usr/src/app/backend

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./target/backend ./src

FROM alpine:3.17 AS start

COPY --from=builder /usr/src/app/backend/target/backend /usr/local/bin/backend

CMD ["backend"]
