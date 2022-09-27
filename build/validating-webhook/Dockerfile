FROM golang:1.18.6-bullseye as builder
RUN apt update && apt install -y libsystemd-dev
WORKDIR /src
COPY . /src/
RUN make build-validating-webhook

FROM debian:bullseye
RUN apt update && apt install -y ca-certificates
COPY --from=builder /src/build/validating-webhook/validating-webhook /validating-webhook
CMD ["/validating-webhook"]
