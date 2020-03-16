FROM golang:latest as build

ARG repository
ENV BUILD_PROJECT_PATH=${GOPATH}/src/${repository}
ENV GO111MODULE=on

RUN if [ -z "$repository" ]; then echo "repository arg is null!"; exit 1; else echo "path===${GOPATH}/src/$repository"; fi

ADD . ${GOPATH}/src/${repository}

RUN mkdir -p /app && cd ${BUILD_PROJECT_PATH} && CGO_ENABLED=0 GOOS=linux go build -o /app/main .

FROM alpine:latest as certs
RUN apk --update add ca-certificates && \
        mkdir -p /app

COPY --from=build /app/main /app/main

WORKDIR /app
CMD ["/app/main"]
EXPOSE 8080
