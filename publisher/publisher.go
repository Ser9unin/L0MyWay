package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Ser9unin/L0MyWay/pkg/model"
	"github.com/bxcodec/faker/v4"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx := context.Background()

	NatsStreamPublisher(ctx)
}
func NatsStreamPublisher(ctx context.Context) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	jets, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	_, err = jets.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:        "Orders",
		Description: "Order publisher",
		Subjects:    []string{"Orders.>"},
	})

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 500; i++ {
		newMsg := publishMessage()
		time.Sleep(200 * time.Millisecond)
		_, err = jets.Publish(ctx, fmt.Sprintf("Orders.%d", i), []byte(newMsg))
		if err != nil {
			log.Fatal()
		}
		log.Printf("Published message %d", i)
	}
}

func publishMessage() []byte {
	fakeOrder := new(model.Model)
	err := faker.FakeData(fakeOrder)
	if err != nil {
		log.Printf("can't create fake data: %s", err.Error())
	}
	byteMsg, err := json.Marshal(fakeOrder)
	if err != nil {
		log.Printf("error marshaling message %s", err.Error())
	}

	if rand.Float64() > 0.9 {
		byteMsg = []byte("bad data")
		fakeOrder.OrderUID = "bad data"
	}

	return byteMsg
}
