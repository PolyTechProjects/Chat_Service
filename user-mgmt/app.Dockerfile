FROM golang:1.22
COPY ./build/UserMgmtApp-* /app/UserMgmtApp
COPY ./.env /app/.env
WORKDIR /app
RUN chmod +x ./UserMgmtApp
ENTRYPOINT [ "./UserMgmtApp" ]
