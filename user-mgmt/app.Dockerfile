FROM golang:1.22
COPY ./build/UserMgmtApp-* /app/UserMgmtApp
COPY ./.env /app/.env
WORKDIR /app
ENTRYPOINT [ "./UserMgmtApp" ]
