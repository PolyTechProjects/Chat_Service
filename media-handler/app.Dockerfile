FROM golang:1.22
COPY ./build/MediaHandlerApp-* /app/MediaHandlerApp
COPY ./.env /app/.env
WORKDIR /app
ENTRYPOINT [ "./MediaHandlerApp" ]