#-------------- Build ----------
FROM golang:1.19-alpine AS build
RUN apk add build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY ./class/* ./class/
COPY ./fonts/ ./fonts/
#COPY ./scripts/* ./scripts/
COPY ./*.go ./
COPY ./*.yaml ./
COPY ./*.png ./

#ENV HTTP_PROXY 
#ENV HTTPS_PROXY 
#ENV http_proxy 
#ENV https_proxy

#RUN CGO_ENABLED=0 go build -v -o ./restserver
RUN go build -v -o ./restserver

#-------------- Deploy ----------
FROM alpine:3
RUN apk add --no-cache sqlite

WORKDIR /rest-app

RUN mkdir images
RUN chown -R 1001:1001 /rest-app

USER 1001:1001

COPY ./fonts/ ./fonts/
COPY ./images/favicon.ico ./images/favicon.ico
COPY ./scripts/build_database.sh ./scripts/build_database.sh
COPY ./scripts/schema.sql ./scripts/schema.sql
COPY ./assets/index.html ./index.html

RUN ./scripts/build_database.sh
RUN sqlite3 "./db/wordCount.db" < ./scripts/schema.sql

COPY --from=build /app/restserver ./restserver

ARG NEW_LISTEN_PORT=9090
ENV LISTEN_PORT=$NEW_LISTEN_PORT

EXPOSE $LISTEN_PORT

CMD ["./restserver"]