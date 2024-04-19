How to run the service
Launch nats and postgres containers:

make run

make migrate

Launch app:

go run ./cmd

Launch publisher (in another terminal):

go run ./publisher

To access API go to 127.0.0.1:8080.
