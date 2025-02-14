FROM alpine:3.21.3

RUN apk add --no-cache ca-certificates

ADD ./cluster-operator /cluster-operator

ENTRYPOINT ["/cluster-operator"]
