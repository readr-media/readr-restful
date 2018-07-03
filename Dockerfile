FROM golang:1.8.5-alpine

WORKDIR /
RUN export GIN_MODE=release

RUN mkdir /config
ADD app /app
ADD db_schema /db_schema

ADD default/readr-restful/main.json /config/main.json
ADD default/readr-restful/newsletter.html /config/newsletter.html

VOLUME /var/log
EXPOSE 8080

CMD ["./app"]
