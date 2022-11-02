FROM alpine:3.16.2

RUN apk add --no-cache ca-certificates

ADD ./cluster-operator /cluster-operator

ENTRYPOINT ["/cluster-operator"]
