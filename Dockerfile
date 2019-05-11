FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ /etc/ssl/certs/
WORKDIR /app
ADD ./bin/ /app/
ADD ./config.yaml /app/
#ADD ./static/ /app/static/
CMD ["/app/main"]
EXPOSE 8080 18080 6060
