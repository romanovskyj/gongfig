#!/bin/bash

if [[ $TRAVIS_OS_NAME == 'linux' ]]; then

    # Push new image
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    docker build -t eromanovskyj/gongfig:latest .
    docker push eromanovskyj/gongfig:latest

fi