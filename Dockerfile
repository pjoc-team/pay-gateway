FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM golang:1.12
COPY --from=certs /etc/ssl/certs/ /etc/ssl/certs/

ENV GO111MODULE=on
ADD . /go/src/gitlab.com/pjoc-team/pay-gateway
ADD config.yaml /app/

RUN echo "path===${GOPATH}/src/$CI_PROJECT_PATH"

RUN mkdir -p /app && cd /go/src/gitlab.com/pjoc-team/pay-gateway && CGO_ENABLED=0 GOOS=linux go build -o /app/main .

WORKDIR /app
#ADD ./bin/ /app/
CMD ["/app/main"]
EXPOSE 5000
