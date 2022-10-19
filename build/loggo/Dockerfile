FROM golang:1.18.6-bullseye as builder
RUN apt update && apt install -y libsystemd-dev
WORKDIR /src
COPY . /src/
RUN make build

FROM debian:bullseye
RUN apt update && apt install -y ca-certificates
COPY --from=builder /src/build/loggo/loggo /loggo
CMD ["/loggo"]
