# Golang Zocket Assignment

## Instructions to run the program:

Install all the modules and dependencies:-
```bash
go get -d
go install 
```

Create a RabbitMQ instance in docker exposing ports for the Management Console and the Message Queues:
```bash 
docker run -d --hostname my-rabbit --name some-rabbit -p 15672:15672 -p 5672:5672 rabbitmq:3-management
```

The credentials for the management platform is:
username: guest
password: guest

Run the API using the following command:
```bash 
go run app.go 
```

Run the consumer using the following command:
```bash 
go run consumer.go
```
