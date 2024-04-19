package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ser9unin/L0MyWay/pkg/api"
	storage "github.com/Ser9unin/L0MyWay/pkg/storage/db"
	cache "github.com/Ser9unin/L0MyWay/pkg/storage/inmemory"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	ctx := context.Background()

	// init config
	if err := config(); err != nil {
		log.Fatal("error config initialisation", err)
	}

	dbcfg := makeStorageCfg()

	// open DB
	db, err := storage.NewDBCreateAndConnect(ctx, dbcfg)
	if err != nil {
		log.Fatal("unable to start db", err)
	}
	defer db.Close()

	// run consumer

	nc, err := nats.Connect(nats.DefaultURL, nats.Name("Orders Consumer"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	jets, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	stream, err := jets.Stream(ctx, "Orders")
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:      "Orders_consumer",
		Durable:   "Orders_consumer",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal(err)
	}

	var dbOrder storage.DBOrder

	memcache, err := cache.NewCache(1024, db)
	if err != nil {
		log.Println("can't create cache", err)
	}

	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		fmt.Println("consumer launched")
		NatsMsgData := msg.Data()

		dbOrder = storage.UnmarshallNatsMsg(NatsMsgData)
		err = storage.InsertToDB(ctx, db, dbOrder)
		memcache.Store(dbOrder.OrderUID, dbOrder.Data)

		if err != nil {
			log.Println("can't insert data to db", err)
		}

		fmt.Println("Recieved: ", string(dbOrder.OrderUID))
		msg.Ack()
	})
	if err != nil {
		log.Fatal(err)
	}

	defer cctx.Stop()

	api := api.API{}
	router := api.NewRouter(db, memcache)
	apiServerAddres := viper.GetString("server.addres")

	apiServer := &http.Server{
		Addr:    apiServerAddres,
		Handler: router,
		// ReadTimeout: time.Duration(config.ReadTimeout) * time.Second,
		// IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
	}

	logger.Info("running http server")
	go func() {
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("can't start server", zap.Error(err), zap.String("server address", apiServerAddres))
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// logger.Info("received an interrupt, closing stan connection and stopping server")

	// timeout, cancel := context.WithTimeout(context.Background(), time.Duration(config.ShutdownTimeout)*time.Second)
	// defer cancel()
	// if err := apiServer.Shutdown(timeout); err != nil {
	// 	logger.Error("can't shutdown http server", zap.Error(err))
	// }
}

func config() error {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("configs")
	return viper.ReadInConfig()
}

func makeStorageCfg() storage.DBConfig {
	return storage.DBConfig{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: viper.GetString("db.password"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	}
}
