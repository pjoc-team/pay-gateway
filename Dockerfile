
FROM scratch
EXPOSE 8080
ENTRYPOINT ["/pay-gateway"]
COPY ./bin/ /
