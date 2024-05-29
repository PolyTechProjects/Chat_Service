docker compose up redis -d
cd auth/ && docker compose up auth_db -d && cd ../
cd user-mgmt/ && docker compose up user_mgmt_db -d && cd ../
cd chat-app/ && docker compose up chat_db -d && cd ../
cd media-handler/ && docker compose up media_db seaweedfs_master seaweedfs_volume1 seaweedfs_volume2 -d && cd ../
cd notification/ && docker compose up notification_db && cd ../
echo '----------------------------------'
echo 'AUTH_POSTGRES' $(docker ps | grep -i "auth_db" | awk '{print $1}')
echo 'USER_MGMT_POSTGRES' $(docker ps | grep -i "user_mgmt_db" | awk '{print $1}')
echo 'CHAT_POSTGRES' $(docker ps | grep -i "chat_db" | awk '{print $1}')
echo 'MEDIA_HANDLER_POSTGRES' $(docker ps | grep -i "media_db" | awk '{print $1}')
echo 'MEDIA_HANDLER_SEAWEEDFS_MASTER' $(docker ps | grep -i "seaweedfs_master" | awk '{print $1}')
echo 'MEDIA_HANDLER_SEAWEEDFS_VOLUME_1' $(docker ps | grep -i "seaweedfs_volume1" | awk '{print $1}')
echo 'MEDIA_HANDLER_SEAWEEDFS_VOLUME_2' $(docker ps | grep -i "seaweedfs_volume2" | awk '{print $1}')
echo 'NOTIFICATION_POSTGRES' $(docker ps | grep -i "notification_db" | awk '{print $1}')
echo 'REDIS' $(docker ps | grep -i "redis" | awk '{print $1}')
echo '----------------------------------'
