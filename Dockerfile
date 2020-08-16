FROM golang:1.14-alpine AS builder

RUN mkdir -p /go/h2-proxy && \
apk update && apk add --no-cache git

WORKDIR /go/h2-proxy

COPY . .

RUN go build ./...  && go build -o ./h2-proxy main.go

FROM alpine

LABEL maintaner="Andres Perez"
LABEL language="golang"
LABEL os="linux"

RUN apk --no-cache add ca-certificates

WORKDIR /

RUN mkdir -p /etc/h2-proxy

COPY --from=builder /go/h2-proxy/h2-proxy /

EXPOSE 8080 8090 50060 50061

CMD [ "/h2-proxy" ]
