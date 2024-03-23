FROM ubuntu:22.04

WORKDIR /

COPY ./bin/simple_endpoint /usr/local/bin/simple_endpoint

WORKDIR /

ENTRYPOINT []
CMD []
