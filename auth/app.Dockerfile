FROM golang:1.22
COPY ./build/AuthApp-* /app/AuthApp
COPY ./.env /app/.env
WORKDIR /app
RUN chmod +x ./AuthApp
ENTRYPOINT [ "./AuthApp" ]