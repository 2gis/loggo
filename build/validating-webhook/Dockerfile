FROM golang:1.27.1-trixie AS builder
RUN apt update && apt install -y libsystemd-dev
WORKDIR /src
COPY . /src/
RUN make build-validating-webhook

FROM debian:trixie
RUN apt update && apt install -y ca-certificates
COPY --from=builder /src/build/validating-webhook/validating-webhook /validating-webhook
ENTRYPOINT ["/validating-webhook"]
