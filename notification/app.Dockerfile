FROM golang:1.22
COPY ./build/NotificationApp-* /app/NotificationApp
COPY ./.env /app/.env
COPY ./firebaseServiceAccount.json /app/firebaseServiceAccount.json
WORKDIR /app
RUN chmod +x ./NotificationApp
ENTRYPOINT [ "./NotificationApp" ]
