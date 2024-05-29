cd auth/ && bash tools/protogen.sh && echo "---" && cd ../
cd user-mgmt/ && bash tools/protogen.sh && echo "---" && cd ../
cd channel-management && bash tools/protogen.sh && echo "---" && cd ../
cd chat-management/ && bash tools/protogen.sh && echo "---" && cd ../
cd chat-app/ && bash tools/protogen.sh && echo "---" && cd ../
cd media-handler/ && bash tools/protogen.sh && echo "---" && cd ../