package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/golang-kafka-3node/internal/kafkautil"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func usage() {
	fmt.Println(`Kafka admin CLI

Usage:
  go run ./cmd/admin create-topic   -topic orders -partitions 6 -replication 3
  go run ./cmd/admin add-partitions -topic orders -add 3
  go run ./cmd/admin set-partitions -topic orders -partitions 12`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	client, err := kgo.NewClient(kgo.SeedBrokers(kafkautil.Brokers()...))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin := kadm.NewClient(client)

	switch os.Args[1] {
	case "create-topic":
		fs := flag.NewFlagSet("create-topic", flag.ExitOnError)
		topic := fs.String("topic", "", "topic name")
		partitions := fs.Int("partitions", 3, "number of partitions")
		replication := fs.Int("replication", 3, "replication factor")
		_ = fs.Parse(os.Args[2:])
		requireTopic(*topic)
		resp, err := admin.CreateTopic(ctx, int32(*partitions), int16(*replication), nil, *topic)
		if err != nil {
			log.Fatalf("create topic: %v", err)
		}
		fmt.Printf("created topic=%s partitions=%d replication=%d\n", resp.Topic, *partitions, *replication)
	case "add-partitions":
		fs := flag.NewFlagSet("add-partitions", flag.ExitOnError)
		topic := fs.String("topic", "", "topic name")
		add := fs.Int("add", 1, "number of partitions to add")
		_ = fs.Parse(os.Args[2:])
		requireTopic(*topic)
		responses, err := admin.CreatePartitions(ctx, *add, *topic)
		if err != nil {
			log.Fatalf("add partitions request: %v", err)
		}
		if err := responses.Error(); err != nil {
			log.Fatalf("add partitions: %v", err)
		}
		fmt.Printf("added %d partition(s) to topic=%s\n", *add, *topic)
	case "set-partitions":
		fs := flag.NewFlagSet("set-partitions", flag.ExitOnError)
		topic := fs.String("topic", "", "topic name")
		partitions := fs.Int("partitions", 0, "final total partition count")
		_ = fs.Parse(os.Args[2:])
		requireTopic(*topic)
		if *partitions <= 0 {
			log.Fatal("-partitions must be > 0")
		}
		responses, err := admin.UpdatePartitions(ctx, *partitions, *topic)
		if err != nil {
			log.Fatalf("set partitions request: %v", err)
		}
		if err := responses.Error(); err != nil {
			log.Fatalf("set partitions: %v", err)
		}
		fmt.Printf("topic=%s requested partition count=%d\n", *topic, *partitions)
	default:
		usage()
		os.Exit(2)
	}
}

func requireTopic(topic string) {
	if topic == "" {
		log.Fatal("-topic is required")
	}
}
