# This project is mini microservice

---



***The purpose of this project*** is to run in the background using redis and rabbitmq to receive and distribute messages and finally show the status.

It contains 4 small programs.

## Working logic:

    The first program generates random order id and gives it to rabbitmq. rabbitmq gives that order id to the 2nd program. the second program prints the order id and status to redis and passes the same order id to the 2nd queue. The 3rd and 4th program repeats the same process. the system stops when the last 4th program presses the status redise.

    Usage guide: You can run the system with a small script called ./start.sh in the repo. You have 2 different ways to test.

1st way:
via cli :
With `curl -X POST http://localhost:8787/order` you generate the order id and activate the program and if you want to see the status of the order id. 

`curl -X POST -d '{"orderid": "8227603"}' http://localhost:8787/status`. Of course you need to change the order id.

2nd way:

there is a python script inside the repo called `request.py` you can use it. it is set to show the current status every 10 seconds on cli.

<mark>TODO:</mark>

I couldn't run it in the background with docker. I get rabbitmq connection refusal error. 


