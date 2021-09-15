FROM golang:1.16.8-bullseye as builder
RUN apt update && apt install -y libsystemd-dev
WORKDIR /src
COPY . /src/
RUN make build

FROM debian:bullseye
COPY --from=builder /src/build/loggo /loggo
CMD ["/loggo"]
