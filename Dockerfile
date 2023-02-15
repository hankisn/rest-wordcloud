#-------------- Build ----------
FROM golang:1.19-alpine AS build
RUN apk add build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY ./class/* ./class/
COPY ./fonts/ ./fonts/
COPY ./scripts/* ./scripts/
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
COPY --from=build /app/scripts/build_database.sh ./scripts/build_database.sh
COPY --from=build /app/scripts/schema.sql ./scripts/schema.sql

RUN ./scripts/build_database.sh
RUN sqlite3 "./db/wordCount.db" < ./scripts/schema.sql

COPY --from=build /app/restserver ./restserver

EXPOSE 80

CMD ["./restserver"]