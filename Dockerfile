FROM golang:1.16 as builder

ADD . /app/
WORKDIR /app/
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/op_exporter /usr/local/bin/
ENTRYPOINT ["op_exporter"]
CMD ["--help"]
