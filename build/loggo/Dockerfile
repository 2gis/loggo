FROM golang:1.27.1-trixie AS builder
RUN apt update && apt install -y libsystemd-dev
WORKDIR /src
COPY . /src/
RUN make build

FROM debian:trixie
RUN apt update && apt install -y ca-certificates
COPY --from=builder /src/build/loggo/loggo /loggo
ENTRYPOINT ["/loggo"]
