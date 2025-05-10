docker compose up redis -d
cd auth/ && docker compose up auth_db -d && cd ../
cd user-mgmt/ && docker compose up user_mgmt_db -d && cd ../
cd chat/ && docker compose up chat_db -d && cd ../
cd messaging/ && docker compose up messaging_db -d && cd ../
cd media-handler/ && docker compose up media_db seaweedfs_master seaweedfs_volume1 seaweedfs_volume2 -d && cd ../
cd notification/ && docker compose up notification_db -d && cd ../
