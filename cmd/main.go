package main

import (
	"context"
	"fmt"
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
		logger.Fatal("error config initialisation: ", zap.Error(err))
	}

	dbcfg := makeStorageCfg()

	// open DB
	db, err := storage.NewDBCreateAndConnect(ctx, dbcfg)
	if err != nil {
		logger.Fatal("unable to start db: ", zap.Error(err))
	}
	defer db.Close()

	// create inmemory cache
	memcache, err := cache.NewCache(1024, db)
	if err != nil {
		logger.Warn("can't create cache: ", zap.Error(err))
	}

	// run consumer
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("Orders Consumer"))
	if err != nil {
		logger.Fatal("unable to start nats connect: ", zap.Error(err))
	}
	defer nc.Close()

	jets, err := jetstream.New(nc)
	if err != nil {
		logger.Fatal("unable to create Jetstream: ", zap.Error(err))
	}

	stream, err := jets.Stream(ctx, "Orders")
	if err != nil {
		logger.Warn("unable to get stream: ", zap.Error(err))
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "Orders_consumer",
	})
	if err != nil {
		logger.Fatal("unable to create or update consumer: ", zap.Error(err))
	}

	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		fmt.Println("consumer launched")
		NatsMsgData := msg.Data()

		dbOrder, err := storage.UnmarshallNatsMsg(NatsMsgData)
		if err != nil {
			logger.Warn("Unable to unmarshall data", zap.Error(err))
		}

		err = storage.InsertToDB(ctx, db, dbOrder)
		if err != nil {
			logger.Warn("can't insert data to db", zap.Error(err))
		}

		memcache.Store(dbOrder.OrderUID, dbOrder.Data)

		fmt.Println("Recieved: ", string(dbOrder.OrderUID))
		msg.Ack()
	})
	if err != nil {
		logger.Fatal("unable to consume message: ", zap.Error(err))
	}

	defer cctx.Stop()

	api := api.API{}
	router := api.NewRouter(db, memcache)
	apiServerAddres := viper.GetString("server.addres")

	apiServer := &http.Server{
		Addr:    apiServerAddres,
		Handler: router,
	}

	logger.Info("running http server")
	if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("can't start server", zap.Error(err), zap.String("server address", apiServerAddres))
	}

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
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
