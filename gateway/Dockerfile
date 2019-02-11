FROM golang:1.8.3 as builder

RUN mkdir -p /go/src/corpochain
RUN curl https://glide.sh/get | sh
ADD . /go/src/corpochain

WORKDIR /go/src/corpochain/gateway
#RUN for p in $(glide nv); do go tool vet "${p%/...}"; done
RUN CGO_ENABLED=0 go install -a corpochain/gateway

FROM alpine:3.6
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /go/bin/gateway .
CMD ["./gateway"]
EXPOSE 80
