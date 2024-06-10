for dir in ./proto/*;
do
    for code_dir in $(ls -d */ | cut -f1 -d'/' | grep -iv "proto");
    do
        mkdir -p ./$code_dir/src/gen/go
        protoc -I=./proto/ $dir/*.proto --go_out=./$code_dir/src/gen/go --go_opt=paths=source_relative --go-grpc_out=./$code_dir/src/gen/go --go-grpc_opt=paths=source_relative;
        echo "Generating code for $dir in $code_dir"
    done
done