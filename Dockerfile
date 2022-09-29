FROM golang:1.18-alpine as builder
WORKDIR $GOPATH/src/go.k6.io/k6
ADD . .
RUN apk --no-cache add build-base git
RUN go install go.k6.io/xk6/cmd/xk6@latest
RUN CGO_ENABLED=1 xk6 build \
    --with github.com/rverma-nsl/xk6-cassandra=. \
    --output /tmp/k6

# Create image for running your customized k6
FROM alpine:latest
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 12345 -g 12345 k6
COPY --from=builder /tmp/k6 /usr/bin/k6

USER 12345
WORKDIR /home/k6

ENTRYPOINT ["k6"]