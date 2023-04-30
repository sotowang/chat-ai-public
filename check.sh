#!/bin/bash

SERVICE_NAME="chat-ai"
while true
do
    if pgrep -x "$SERVICE_NAME" >/dev/null
    then
        echo "$SERVICE_NAME is running"
    else
        echo "$SERVICE_NAME stopped"
        # restart service
        nohup /root/sotowang/chat-ai/chat-ai --config=/root/sotowang/chat-ai/config-prod.ini  > output.log 2>&1 &

    fi
    sleep 15s # check every 5 seconds
done