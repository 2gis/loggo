FROM ubuntu:20.04
COPY ./build/loggo /loggo
CMD ["/loggo"]
