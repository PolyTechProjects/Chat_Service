cd auth/ && docker compose stop auth_app && cd ../
cd user-mgmt/ && docker compose stop user_mgmt_app && cd ../
cd chat/ && docker compose stop chat_app && cd ../
cd messaging/ && docker compose stop messaging_app && cd ../
cd media-handler/ && docker compose stop media_app && cd ../
cd notification/ && docker compose stop notification_app && cd ../
docker ps
