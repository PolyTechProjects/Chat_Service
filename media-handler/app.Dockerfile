FROM golang:1.22
COPY ./build/MediaHandlerApp-* /app/MediaHandlerApp
COPY ./config/local.yaml /app/config/config.yaml
COPY ./.env /app/.env
WORKDIR /app
ENV CONFIG_PATH="./config/config.yaml"
ENTRYPOINT [ "./MediaHandlerApp" ]