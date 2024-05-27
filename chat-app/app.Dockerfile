FROM golang:1.22
COPY ./build/ChatApp-* /app/ChatApp
COPY ./.env /app/.env
WORKDIR /app
ENV CONFIG_PATH="./config/config.yaml"
ENTRYPOINT [ "./ChatApp" ]
