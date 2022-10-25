FROM golang:1.19.2-alpine3.16 as build
WORKDIR /server
COPY . /server
RUN go build -o /server-app

FROM alpine
COPY --from=build ./server-app ./
COPY --from=build ./server/views ./views
COPY --from=build ./server/templates ./templates
COPY --from=build ./server/.env ./
EXPOSE 3001
ENTRYPOINT ["/server-app"]