FROM golang:1.22
COPY ./build/ChannelManagementService-* /app/ChannelManagementService
COPY ./.env /app/.env
WORKDIR /app
ENTRYPOINT [ "./ChannelManagementService" ]