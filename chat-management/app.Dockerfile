FROM golang:1.22
COPY ./build/ChatManagementService-* /app/ChatManagementService
COPY ./.env /app/.env
WORKDIR /app
ENTRYPOINT [ "./ChatManagementService" ]