FROM golang

RUN apt-get update
RUN apt-get -y upgrade
RUN apt-get -y update
RUN apt-get -y install \
    build-essential \
    wget \
    curl \
    git \
    make \
    unzip

ENV GO111MODULE=on

RUN GO111MODULE=off go get -u github.com/myitcv/gobin

ENV GOLANG_PROTOBUF_VERSION=1.3.1
## install protoc-gen-go, sqlboiler, sqlboiler-psql
RUN gobin github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION}
RUN gobin github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN gobin github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN gobin github.com/mwitkow/go-proto-validators/protoc-gen-govalidators

ENV PROTOBUF_VERSION=3.6.1

RUN curl -sSL https://github.com/uber/prototool/releases/download/v1.9.0/prototool-$(uname -s)-$(uname -m) \
    -o /usr/local/bin/prototool && \
    chmod +x /usr/local/bin/prototool

## install protoc
RUN mkdir -p /tmp/protoc && \
  curl -sSL \
  https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip \
  -o /tmp/protoc/protoc.zip && \
  cd /tmp/protoc && \
  unzip protoc.zip && \
  mv /tmp/protoc/include /usr/local/include && mv /tmp/protoc/bin/protoc /usr/local/bin/protoc

RUN git clone https://github.com/googleapis/googleapis
RUN mv googleapis/google/ /usr/local/include/google

RUN mv /go/bin/* /usr/local/bin/
RUN chmod +x /usr/local/bin/*
WORKDIR /tmp
RUN rm -rf protoc