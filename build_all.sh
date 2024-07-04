for dir in $(ls -d */ | cut -f1 -d'/' | grep -iv "proto");
do
    echo "----------------------------------"
    echo "Building $dir"
    cd $dir
    if bash build.sh; then
        echo "Build successful"
    else
        echo "Build failed"
        break
    fi
    service=$(docker compose config --services | grep -i "app")
    if docker compose build $service; then
        echo "Docker Compose build successful"
    else
        echo "Docker Compose build failed"
        break
    fi
    cd ../
done
echo "----------------------------------"