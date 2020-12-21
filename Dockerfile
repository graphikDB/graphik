FROM golang:1.14.2-alpine3.11 as build-env

RUN mkdir /graphik
RUN apk --update add ca-certificates
RUN apk add make git
WORKDIR /graphik
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go install ./...

FROM alpine
RUN apk add ca-certificates
COPY --from=build-env /go/bin/ /usr/local/bin/
WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/graphik"]