FROM golang:1.10

RUN go get github.com/romanovskyj/gongfig

CMD ["tail", "-f", "/dev/null"]