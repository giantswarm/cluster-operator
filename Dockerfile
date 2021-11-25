FROM alpine:3.15.0

RUN apk add --no-cache ca-certificates

ADD ./cluster-operator /cluster-operator

ENTRYPOINT ["/cluster-operator"]
