package consumer

import (
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	nc, err := nats.Connect("connect.ngs.global", nats.UserCredentials("../user.creds"), nats.Name("Orders Publisher"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	jets, err := jetstream.New(nc)

	if err != nil {
		lof.Fatal(err)
	}

	_, err = jets.AddStream
}
