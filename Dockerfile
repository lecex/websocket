FROM bigrocs/golang-gcc:1.13 as builder

ARG ACCES_STOKEN
RUN apk add git
RUN go env -w GOPRIVATE=github.com/lecex,github.com/bigrocs
RUN git config --global url."https://bigrocs:${ACCES_STOKEN}@github.com".insteadOf "https://github.com"

WORKDIR /go/src/github.com/lecex/websocket
COPY . .

ENV GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN go build -a -installsuffix cgo -o bin/websocket

FROM bigrocs/alpine:ca-data

RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /go/src/github.com/lecex/websocket/bin/websocket /usr/local/bin/
CMD ["websocket"]
EXPOSE 8080
EXPOSE 8989
