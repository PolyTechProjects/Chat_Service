FROM golang:1.22
COPY ./build/NotificationApp-* /app/NotificationApp
COPY ./.env /app/.env
WORKDIR /app
ENTRYPOINT [ "./NotificationApp" ]
