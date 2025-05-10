FROM golang:1.22
COPY ./build/UsersApp-* /app/UsersApp
COPY ./.env /app/.env
WORKDIR /app
RUN chmod +x ./UsersApp
ENTRYPOINT [ "./UsersApp" ]
