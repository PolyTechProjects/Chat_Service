FROM golang:1.22
COPY ./build/ChatApp-* /app/ChatApp
COPY ./.env /app/.env
WORKDIR /app
RUN chmod +x ./ChatApp
ENTRYPOINT [ "./ChatApp" ]
