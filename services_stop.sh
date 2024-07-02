cd auth/ && docker compose stop auth_app && cd ../
cd user-mgmt/ && docker compose stop user_mgmt_app && cd ../
cd channel-management/ && docker compose stop channel_management_app && cd ../
cd chat-management/ && docker compose stop chat_management_app && cd ../
cd chat-app/ && docker compose stop chat_app && cd ../
cd media-handler/ && docker compose stop media_app && cd ../
cd notification/ && docker compose stop notification_app && cd ../
docker ps
