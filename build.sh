repository=`cat go.mod | grep -E "^module\s[0-9a-zA-Z\./_\-]+" | awk '{print $2}'`
docker build --build-arg repository=$repository . -t image --file Dockerfile
