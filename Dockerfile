FROM golang:1.8.5-alpine

WORKDIR /
RUN export GIN_MODE=release

ADD app /app
ADD db_schema /db_schema

ADD config /config

VOLUME /var/log
EXPOSE 8080

CMD ["./app"]
