#!/bin/bash

# This document is Licensed under Creative Commons CC0.
# To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
# to this document to the public domain worldwide.
# This document is distributed without any warranty.
# You should have received a copy of the CC0 Public Domain Dedication along with this document.
# If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

# Make sure you set $DATAGET and $DATASET

# export DATASET=https://example.com?apikey=abcd
# export DATAGET=https://example.com

printf "" | curl -X PUT --data-binary @- $DATASET'&format=%25s' || echo Environment not set.
(curl $DATAGET/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.tig && echo OK.) || echo Environment not set.

echo >/tmp/tig.sh
test -f $IMPLEMENTATION/private.key && cat $IMPLEMENTATION/private.key|curl -X PUT --data-binary @- $DATASET'&format=curl%20'$DATAGET'*%20>/data/key.pem' >>/tmp/tig.sh
echo >>/tmp/tig.sh
test -f $IMPLEMENTATION/certificate.crt && test -f $IMPLEMENTATION/ca_bundle.crt && cat $IMPLEMENTATION/certificate.crt $IMPLEMENTATION/ca_bundle.crt | curl -X PUT --data-binary @- $DATASET'&format=curl%20'$DATAGET'*%20>/data/certificate.pem' >>/tmp/tig.sh
echo >>/tmp/tig.sh
echo cd /go/src >>/tmp/tig.sh
echo >>/tmp/tig.sh
tar --exclude .git -c --exclude='documentation/benchmark1' --exclude='documentation/benchmark2' . | curl --data-binary @- -X POST $DATASET'&format=''*' >/tmp/codepath
echo curl $DATAGET$(cat /tmp/codepath)" | tar -x" >>/tmp/tig.sh
echo go run main.go >>/tmp/tig.sh
cat /tmp/tig.sh | curl -X PUT --data-binary @- $DATASET'&format=*' >/tmp/tig.txt

# This is a bit invasive. Should we check for build changes?
echo >/tmp/tig.sh
echo 'apt update && apt install -y docker.io && service docker start && docker ps' >>/tmp/tig.sh
echo 'yum install -y docker  && docker ps && touch /etc/containers/nodocker && service docker start' >>/tmp/tig.sh
echo docker stop storm >>/tmp/tig.sh
echo docker rm storm >>/tmp/tig.sh
echo docker run --name storm -d --restart=always -p 443:443 --tmpfs /data:rw,size=4g docker.io/library/golang@sha256:10e3c0f39f8e237baa5b66c5295c578cac42a99536cc9333d8505324a82407d9 bash -c \''curl '$DATAGET$(cat /tmp/tig.txt)'|bash'\' >>/tmp/tig.sh
echo

echo
echo Code file
echo curl $DATAGET$(cat /tmp/codepath)" | tar -t"
echo
echo Docker launch file
cat /tmp/tig.sh | curl -X PUT --data-binary @- $DATASET'&format=curl%20'$DATAGET'*'%20%7C%20sudo%20bash
#printf "curl $DATAGET$(cat /tmp/codepath) | sudo bash"


