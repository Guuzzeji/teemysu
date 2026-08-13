# syntax=docker/dockerfile:1

# ---- Build stage -----------------------------------------------------------
# mattn/go-sqlite3 and sqlite-vec are cgo packages: needs C compiler + musl
# headers so the binary links cleanly against alpine's musl libc.
FROM golang:1.26.5-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /src

# Cache dependencies first: only re-run `go mod download` when go.mod changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# sqlite_fts5 matches the makefile; single static-ish binary, stripped.
# sqlite-vec.c assumes glibc's u_int*_t typedefs, missing on musl; alias them.
RUN CGO_CFLAGS="-Du_int8_t=uint8_t -Du_int16_t=uint16_t -Du_int32_t=uint32_t -Du_int64_t=uint64_t" \
    CGO_ENABLED=1 go build -tags sqlite_fts5 -ldflags '-s -w' -o /out/bot .

# ---- Runtime stage ---------------------------------------------------------
FROM alpine:3.21

# CA certs for Discord/OpenAI TLS; non-root user for the container.
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 10001 app

WORKDIR /app
COPY --from=builder /out/bot .

# Mount point for the SQLite database (named volume in docker-compose).
RUN mkdir -p /data && chown app:app /data

USER app

ENV DATABASE_PATH=/data/data.db
VOLUME ["/data"]

ENTRYPOINT ["/app/bot"]