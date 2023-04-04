#!/bin/bash

# Start Redis container
docker rm -f redis && docker run -d --name redis -p 6379:6379 redis

# Start RabbitMQ container
docker rm -f rabbitmq && docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# Wait for Redis and RabbitMQ to start
sleep 5

# Start the Go programs
server/main &
sleep 2
setup/setup &
sleep 2
delivering/delivering &
sleep 2
delivered/delivered &

# Wait for the Go programs to start
sleep 5

# Start the request.py script and display output in the CLI

python3  request.py
