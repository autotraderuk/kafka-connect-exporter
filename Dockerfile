FROM golang:1.11 as builder
WORKDIR /go/src/github.com/snahelou/kafka-connect-exporter
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo . 

FROM alpine:3.8
COPY --from=builder /go/src/github.com/snahelou/kafka-connect-exporter/kafka-connect-exporter /
ENTRYPOINT ["/kafka-connect-exporter"]
