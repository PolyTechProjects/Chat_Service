cd auth/ && docker compose stop auth_db && cd ../
cd user-mgmt/ && docker compose stop user_mgmt_db && cd ../
cd chat-app/ && docker compose stop chat_db && cd ../
cd media-handler/ && docker compose stop media_db seaweedfs_master seaweedfs_volume1 seaweedfs_volume2 && cd ../
docker compose stop redis
docker ps
