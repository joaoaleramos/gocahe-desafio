#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Uso: $0 <IP>"
    exit 1
fi

IP=$1


URL="http://localhost:8080"

# Make http request 
curl -X GET "$URL" \
     -H "X-Forwarded-For: $IP" \
     -d "{\"ip\": \"$IP\"}"

