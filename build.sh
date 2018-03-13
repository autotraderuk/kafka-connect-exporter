#!/bin/bash
set -e
ROOT=$(cd "$(dirname "$0")"; pwd)
NAME=$(echo $TRAVIS_REPO_SLUG | sed 's|.*/||')
RELEASE=$(git describe --always --tags)
IMAGE=$TRAVIS_REPO_SLUG

echo "Building Binary ..."
mkdir build dist
gox -output "build/{{.OS}}_{{.Arch}}/{{.Dir}}"
cp build/linux_amd64/${NAME} .

echo "Packaging Binary ..."
cd build
for OSARCH in $(ls); do
    cd "$OSARCH"
    tar -czf "$ROOT/dist/${NAME}_${RELEASE}_${OSARCH}.tar.gz" .
    cd - > /dev/null
done

cd $ROOT

if [ "$RELEASE" != "" ]; then
    echo "Building Docker Image ..."
    docker build -t $IMAGE .
    docker tag ${IMAGE}:latest ${IMAGE}:${RELEASE}

    echo "Pushing Docker Image ..."
    docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"
    docker push ${IMAGE}:${RELEASE}
    docker push ${IMAGE}:latest
else
    echo "Not a release. Not building binaries or pushing image."
fi
