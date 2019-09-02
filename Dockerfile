# Building stage
FROM golang:1.11-alpine3.9

WORKDIR /go/src/github.com/romanovskyj/gongfig

RUN apk add --no-cache ca-certificates git

ENV GO111MODULE=on

COPY go.mod /go/src/github.com/romanovskyj/gongfig/go.mod
COPY go.sum /go/src/github.com/romanovskyj/gongfig/go.sum

RUN go mod vendor -v

ENV CGO_ENABLED 0
ENV GOOS linux

RUN go install ./...

# Production stage
FROM alpine:3.9

WORKDIR /opt

# copy the go binaries from the building stage
COPY --from=0 /go/bin /opt

ENTRYPOINT ["/opt/gongfig"]

CMD ["--help"]
