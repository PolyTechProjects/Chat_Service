cd auth/ && bash tools/protogen.sh && echo "auth" && cd ../
cd user-mgmt/ && bash tools/protogen.sh && echo "user-mgmt" && cd ../
cd channel-management && bash tools/protogen.sh && echo "channel-management" && cd ../
cd chat-management/ && bash tools/protogen.sh && echo "chat-management" && cd ../
cd chat-app/ && bash tools/protogen.sh && echo "chat-app" && cd ../
cd notification/ && bash tools/protogen.sh && echo "notification" && cd ../
cd media-handler/ && bash tools/protogen.sh && echo "media-handler" && cd ../