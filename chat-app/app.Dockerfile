FROM golang:1.22
COPY ./build/ChatApp-* /app/ChatApp
COPY ./.env /app/.env
WORKDIR /app
ENTRYPOINT [ "./ChatApp" ]
