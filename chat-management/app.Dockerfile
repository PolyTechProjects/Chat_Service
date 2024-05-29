FROM golang:1.22
COPY ./build/ChatManagementService-* /app/ChatManagementService
COPY ./config/local.yaml /app/config/config.yaml
COPY ./.env /app/.env
WORKDIR /app
ENV CONFIG_PATH="./config/config.yaml"
ENTRYPOINT [ "./ChatManagementService" ]