package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"example.com/golang-kafka-3node/internal/kafkautil"
	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	topic := flag.String("topic", "orders", "Kafka topic")
	count := flag.Int("count", 10, "number of messages")
	keyPrefix := flag.String("key-prefix", "customer", "message key prefix")
	message := flag.String("message", "hello-kafka", "message prefix")
	delay := flag.Duration("delay", 200*time.Millisecond, "delay between messages")
	flag.Parse()

	client, err := kgo.NewClient(
		kgo.SeedBrokers(kafkautil.Brokers()...),
		kgo.DefaultProduceTopic(*topic),
		kgo.RequiredAcks(kgo.AllISRAcks()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := context.Background()
	for i := 1; i <= *count; i++ {
		key := *keyPrefix + "-" + strconv.Itoa(i%5)
		value := fmt.Sprintf("%s-%d", *message, i)
		record := &kgo.Record{Key: []byte(key), Value: []byte(value), Headers: []kgo.RecordHeader{{Key: "source", Value: []byte("go-producer")}}}
		if err := client.ProduceSync(ctx, record).FirstErr(); err != nil {
			log.Fatalf("produce: %v", err)
		}
		fmt.Printf("produced topic=%s key=%s value=%s partition=%d offset=%d\n", record.Topic, key, value, record.Partition, record.Offset)
		time.Sleep(*delay)
	}
}
