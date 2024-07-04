cd auth/ && docker compose up auth_app -d && cd ../
cd user-mgmt/ && docker compose up user_mgmt_app -d && cd ../
cd channel-management/ && docker compose up channel_management_app -d && cd ../
cd chat-management/ && docker compose up chat_management_app -d && cd ../
cd chat-app/ && docker compose up chat_app -d && cd ../
cd media-handler/ && docker compose up media_app -d && cd ../
cd notification/ && docker compose up notification_app -d && cd ../
echo '----------------------------------'
echo 'AUTH_APP' $(docker ps | grep -i "auth_app" | awk '{print $1}')
echo 'USER_MGMT_APP' $(docker ps | grep -i "user_mgmt_app" | awk '{print $1}')
echo 'CHANNEL_MGMT_APP' $(docker ps | grep -i "channel_management_app" | awk '{print $1}')
echo 'CHAT_MGMT_APP' $(docker ps | grep -i "chat_management_app" | awk '{print $1}')
echo 'CHAT_APP' $(docker ps | grep -i "chat_app" | awk '{print $1}')
echo 'MEDIA_HANDLER_APP' $(docker ps | grep -i "media_app" | awk '{print $1}')
echo 'NOTIFICATION_APP' $(docker ps | grep -i "notification_app" | awk '{print $1}')
echo '----------------------------------'
