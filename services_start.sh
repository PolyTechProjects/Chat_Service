cd auth/ && docker compose up auth_app -d && cd ../
cd user-mgmt/ && docker compose up user_mgmt_app -d && cd ../
cd chat/ && docker compose up chat_app -d && cd ../
cd messaging/ && docker compose up messaging_app -d && cd ../
cd media-handler/ && docker compose up media_app -d && cd ../
cd notification/ && docker compose up notification_app -d && cd ../
