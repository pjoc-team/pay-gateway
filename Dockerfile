FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM golang:1.12
COPY --from=certs /etc/ssl/certs/ /etc/ssl/certs/

ENV GO111MODULE=on
ADD . /go/src/gitlab.com/pjoc-team/pay-gateway

RUN echo "path===${GOPATH}/src/$CI_PROJECT_PATH"

RUN mkdir /app && cd /go/src/gitlab.com/pjoc-team/pay-gateway && CGO_ENABLED=0 GOOS=linux go build -o /app/main .


WORKDIR /app
#ADD ./bin/ /app/
ADD /go/src/gitlab.com/pjoc-team/pay-gateway/config.yaml /app/
CMD ["/app/main"]
EXPOSE 5000
