FROM golang:1.22
COPY ./build/MessagingApp-* /app/MessagingApp
COPY ./.env /app/.env
WORKDIR /app
RUN chmod +x ./MessagingApp
ENTRYPOINT [ "./MessagingApp" ]
