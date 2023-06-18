FROM golang:1.19 AS lang

ADD . /opt/app
WORKDIR /opt/app
RUN go build ./cmd/main.go

FROM ubuntu:20.04
RUN apt-get -y update &&\
    apt-get install -y tzdata &&\
    apt-get -y update && apt-get install -y postgresql-12

ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    /etc/init.d/postgresql stop

EXPOSE 5432

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /usr/src/app

COPY . .
COPY --from=lang /opt/app/main .

EXPOSE 5000
ENV PGPASSWORD docker
CMD service postgresql start &&  psql -h localhost -d docker -U docker -p 5432 -a -q -f ./db/db.sql && ./main
