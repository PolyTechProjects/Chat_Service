for dir in proto/*;
do protoc -I proto $dir/*.proto  --go_out=./src/gen/go/ --go_opt=paths=source_relative --go-grpc_out=./src/gen/go/ --go-grpc_opt=paths=source_relative;
done