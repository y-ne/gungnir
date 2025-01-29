FROM rust:1-alpine AS builder

RUN apk add --no-cache musl-dev

WORKDIR /usr/src/gungnir

COPY . .

RUN cargo build --release

FROM alpine:latest

WORKDIR /usr/local/bin

COPY --from=builder /usr/src/gungnir/target/release/gungnir .

EXPOSE 8888

CMD ["./gungnir"]