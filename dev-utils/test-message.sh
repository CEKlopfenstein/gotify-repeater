#!/bin/bash
while true
do
    curl "http://localhost:8080/message?token=$1" -F "title=Test Message" -F "message=Repeating Test Message"
    echo
    if [ $? -ne 0 ]
    then
        break
    fi
sleep 10
done

