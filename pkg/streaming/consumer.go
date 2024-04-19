package server

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func NatsStreamConsumer(ctx context.Context) []byte {
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("Orders Consumer"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	jets, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	stream, err := jets.Stream(ctx, "orders")
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:    "orders_consumer",
		Durable: "orders_consumer",
	})
	if err != nil {
		log.Fatal(err)
	}

	var NatsMsgData []byte
	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		NatsMsgData = msg.Data()
		log.Printf("Recieved: %s", string(msg.Subject()))
		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}

	defer cctx.Stop()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)
	<-quit
	return NatsMsgData
}
