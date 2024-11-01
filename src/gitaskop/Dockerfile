FROM golang:1.19.8 AS builder

WORKDIR /src
COPY *.go go.* /src/

RUN go mod download
RUN go build -o /bin/gitaskop -ldflags "-s -w" .

RUN strip /bin/gitaskop

FROM alpine
RUN apk add --no-cache git openssh-client bash curl jq coreutils ca-certificates
COPY --from=builder /bin/gitaskop /bin/gitaskop