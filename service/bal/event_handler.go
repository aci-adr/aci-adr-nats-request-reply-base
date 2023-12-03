package bal

import (
	"aci-adr-go-base/model/entity"
	"aci-adr-go-base/model/request"
	"aci-adr-go-base/service/dal"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.opentelemetry.io/otel/metric"
)

func Connect(meter metric.Meter) {
	//create DB services as needed.
	var db dal.Database[entity.ForexData] = &dal.MongoDbService[entity.ForexData]{Collection: "forex_data"}
	histogram, _ := meter.Float64Histogram(
		os.Getenv("STAGE_NAME")+"_duration",
		metric.WithDescription("The duration of task execution."),
		metric.WithUnit("s"),
	)

	apiCounter, _ := meter.Int64Counter(
		os.Getenv("STAGE_NAME")+"_total_processed",
		metric.WithDescription("Number of API calls."),
		metric.WithUnit("{call}"),
	)

	nc, conErr := nats.Connect(os.Getenv("NATS_URI"))

	if conErr != nil {
		log.Fatal("Unable to connect to Nats")
	}

	_, err := nc.QueueSubscribe(os.Getenv("LISTEN_SUBJECT"), os.Getenv("GROUP"), func(msg *nats.Msg) {
		go handle(msg, db, histogram, apiCounter)
	})
	if err != nil {
		log.Fatal("Unable to subscribe")
	}

	defer nc.Drain()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	<-signalCh
	log.Println("Shutting down gracefully.")
}

func handle(msg *nats.Msg, db dal.Database[entity.ForexData], histogram metric.Float64Histogram, apiCounter metric.Int64Counter) {
	start := time.Now()
	var fxRequest request.ForexRequest
	unmarshalErr := json.Unmarshal(msg.Data, &fxRequest)
	log.Println("Processing Message:", fxRequest.TenantId)
	if unmarshalErr != nil {
		log.Println("Error in unmarshal.")
		return
	}

	filter := bson.D{
		{Key: "tenantId", Value: fxRequest.TenantId},
		{Key: "bankId", Value: fxRequest.BankId},
		{Key: "baseCurrency", Value: fxRequest.BaseCurrency},
		{Key: "targetCurrency", Value: fxRequest.TargetCurrency},
		{Key: "tier", Value: fxRequest.Tier},
	}

	_, _ = db.GetOne(filter)
	end := time.Now()
	duration := time.Since(start)
	histogram.Record(context.Background(), duration.Seconds())
	apiCounter.Add(context.Background(), 1)
	msg.Respond([]byte("{\"startedOn\": " + strconv.FormatInt(start.UnixMilli(), 10) + ", \"completedOn\": " + strconv.FormatInt(end.UnixMilli(), 10) + "}"))

}
