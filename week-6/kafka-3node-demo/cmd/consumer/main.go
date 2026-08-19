package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/golang-kafka-3node/internal/kafkautil"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	topic := flag.String("topic", "orders", "Kafka topic")
	group := flag.String("group", "orders-workers", "consumer group ID")
	instance := flag.String("instance", "consumer-1", "logical consumer name")
	fromBeginning := flag.Bool("from-beginning", false, "start earliest if group has no committed offset")
	flag.Parse()

	opts := []kgo.Opt{
		kgo.SeedBrokers(kafkautil.Brokers()...),
		kgo.ConsumerGroup(*group),
		kgo.ConsumeTopics(*topic),
		kgo.Balancers(kgo.CooperativeStickyBalancer()),
		kgo.OnPartitionsAssigned(func(_ context.Context, _ *kgo.Client, p map[string][]int32) {
			fmt.Printf("[%s] assigned: %v\n", *instance, p)
		}),
		kgo.OnPartitionsRevoked(func(_ context.Context, _ *kgo.Client, p map[string][]int32) {
			fmt.Printf("[%s] revoked: %v\n", *instance, p)
		}),
	}
	if *fromBeginning {
		opts = append(opts, kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	}
	client, err := kgo.NewClient(opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	fmt.Printf("[%s] consuming topic=%s group=%s brokers=%v\n", *instance, *topic, *group, kafkautil.Brokers())
	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil {
			fmt.Printf("[%s] stopping\n", *instance)
			return
		}
		for _, e := range fetches.Errors() {
			log.Printf("[%s] fetch error topic=%s partition=%d: %v", *instance, e.Topic, e.Partition, e.Err)
		}
		fetches.EachRecord(func(r *kgo.Record) {
			fmt.Printf("[%s] topic=%s partition=%d offset=%d key=%s value=%s\n", *instance, r.Topic, r.Partition, r.Offset, string(r.Key), string(r.Value))
		})
	}
}
