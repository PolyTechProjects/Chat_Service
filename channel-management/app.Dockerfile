FROM golang:1.22
COPY ./build/ChannelManagementService-* /app/ChannelManagementService
COPY ./config/local.yaml /app/config/config.yaml
COPY ./.env /app/.env
WORKDIR /app
ENV CONFIG_PATH="./config/config.yaml"
ENTRYPOINT [ "./ChannelManagementService" ]