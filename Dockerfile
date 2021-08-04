FROM ubuntu:16.04
COPY ./build/loggo /loggo
CMD ["/loggo"]
