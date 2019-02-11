FROM golang:1.8.3 as builder

RUN mkdir -p /go/src/corpochain
RUN curl https://glide.sh/get | sh
ADD . /go/src/corpochain

WORKDIR /go/src/corpochain/corpochain
#RUN for p in $(glide nv); do go tool vet "${p%/...}"; done
RUN CGO_ENABLED=0 go install -a corpochain/corpochain

FROM alpine:3.6
#RUN apk --no-cache add ca-certificates bash
RUN apk --update --no-cache add bash ca-certificates

WORKDIR /root/
COPY --from=builder /go/bin/corpochain .
ENTRYPOINT ["./corpochain", "run"]
EXPOSE 80
