FROM golang:1.22
COPY ./build/NotificationApp-* /app/NotificationApp
COPY ./.env /app/.env
COPY ./google-services.json /app/google-services.json
WORKDIR /app
ENTRYPOINT [ "./NotificationApp" ]
