# Building stage
FROM golang:1.11-alpine3.9

WORKDIR /go/src/github.com/romanovskyj/gongfig

RUN apk add --no-cache ca-certificates git

RUN go get github.com/urfave/cli
RUN go get github.com/jinzhu/copier
RUN go get github.com/mitchellh/mapstructure

COPY pkg /go/src/github.com/romanovskyj/gongfig/pkg
COPY gongfig.go /go/src/github.com/romanovskyj/gongfig

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
