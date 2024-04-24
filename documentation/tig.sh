#!/bin/bash

# Make sure you set $DATAGET and $DATASET

printf "" | curl -X PUT --data-binary @- $DATASET'&format=%25s' || echo Environment not set.
curl $DATAGET/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.tig || echo Environment not set.

export IMPLEMENTATION=$(cat ./res/demo.txt | curl -X PUT --data-binary @- $DATASET'&format='$DATAGET'%25s' || echo Environment not set.)

# http run: IMPLEMENTATION=https://data.schmied.us/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.tig go run main.go
# SSL run: IMPLEMENTATION=https://data.schmied.us/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.tig go run main.go

tar --exclude './.git' -c . | curl -X PUT --data-binary @- $DATASET'&format=https://data.schmied.us%25s' >/tmp/codepath
echo
echo Code file
printf "$(cat /tmp/codepath) | tar -t"
echo
echo Docker launch file
export VERSIONX=7418
printf "docker stop --time 2 storm$VERSIONX;sleep 3;docker run --name storm$VERSIONX -d --rm -e IMPLEMENTATION=$IMPLEMENTATION -e CODE=$(cat /tmp/codepath) -p $VERSIONX:7777 golang:1.19.3 bash -c 'curl \$CODE | tar -x -C /go/src/;cd /go/src/;go run main.go'\n" | curl -X PUT --data-binary @- $DATASET'&format=https://data.schmied.us%25s'
printf " | bash"
echo

