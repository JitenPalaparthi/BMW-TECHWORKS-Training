# Golang + Kafka 3-Node Complete Demo

This project demonstrates a **3-node Apache Kafka 4.3.1 KRaft cluster** and a Go application using **franz-go**.

## What is included

- 3 Kafka brokers/controllers: `kafka1`, `kafka2`, `kafka3`
- No ZooKeeper; KRaft mode
- Go topic administration
- Create a topic with configurable partitions and replication factor
- Increase partitions by an additional count
- Set the final partition count
- Go producer with keys
- Go group consumer
- Multiple consumers in one consumer group
- Multiple independent consumer groups
- Rebalance callbacks showing assigned/revoked partitions
- Commands to inspect topics, replicas, ISR, groups and lag

## Architecture

```text
Go Producer
    |
    v
orders topic: P0 P1 P2 P3 P4 P5 ...
    |
    +---- replicated across Kafka 1 / Kafka 2 / Kafka 3
    |
    v
Consumer Group: orders-workers
    +-- consumer-1
    +-- consumer-2
    +-- consumer-3
```

Within one consumer group, partitions are shared among consumers. A partition is processed by at most one active consumer in that traditional group at a time. A second group such as `analytics-service` consumes the topic independently.

## 1. Start Kafka

```bash
docker compose up -d
docker compose ps
```

External broker addresses used by Go:

```text
localhost:19092
localhost:29092
localhost:39092
```

Optional override:

```bash
export KAFKA_BROKERS=localhost:19092,localhost:29092,localhost:39092
```

## 2. Download Go dependencies

```bash
go mod tidy
```

## 3. Create a topic

Create `orders` with 6 partitions and RF=3:

```bash
go run ./cmd/admin create-topic -topic orders -partitions 6 -replication 3
```

Or:

```bash
make create-topic TOPIC=orders PARTITIONS=6 REPLICATION=3
```

More partitions are allowed at creation time:

```bash
go run ./cmd/admin create-topic -topic payments -partitions 12 -replication 3
```

## 4. Describe the topic

```bash
docker exec kafka1 /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server kafka1:9092 \
  --describe --topic orders
```

This shows partition IDs, leaders, replicas and ISR.

## 5. Increase partitions

Add 3 more:

```bash
go run ./cmd/admin add-partitions -topic orders -add 3
```

Set a final total count:

```bash
go run ./cmd/admin set-partitions -topic orders -partitions 12
```

Kafka partitions can be increased, but not reduced. Increasing partition count should be planned carefully if the application depends on stable key-to-partition mapping.

## 6. Start one consumer

```bash
go run ./cmd/consumer \
  -topic orders \
  -group orders-workers \
  -instance consumer-1 \
  -from-beginning
```

A normal Kafka consumer group does not need to be manually pre-created. The group is established when consumers join with the same group ID and Kafka begins maintaining group membership/offset state.

## 7. Start three consumers in the same group

Terminal 1:

```bash
go run ./cmd/consumer -topic orders -group orders-workers -instance consumer-1 -from-beginning
```

Terminal 2:

```bash
go run ./cmd/consumer -topic orders -group orders-workers -instance consumer-2 -from-beginning
```

Terminal 3:

```bash
go run ./cmd/consumer -topic orders -group orders-workers -instance consumer-3 -from-beginning
```

The application prints rebalance assignments, for example:

```text
[consumer-1] assigned: map[orders:[0 3]]
[consumer-2] assigned: map[orders:[1 4]]
[consumer-3] assigned: map[orders:[2 5]]
```

Exact assignments can differ.

## 8. Publish messages

```bash
go run ./cmd/producer -topic orders -count 30
```

Custom values:

```bash
go run ./cmd/producer \
  -topic orders \
  -count 100 \
  -key-prefix customer \
  -message order-created
```

The output shows the resulting partition and offset.

## 9. Create another consumer group

Run:

```bash
go run ./cmd/consumer \
  -topic orders \
  -group analytics-service \
  -instance analytics-1 \
  -from-beginning
```

Now `orders-workers` and `analytics-service` consume independently.

## 10. List consumer groups

```bash
docker exec kafka1 /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server kafka1:9092 --list
```

Or:

```bash
make groups
```

## 11. Describe group offsets and lag

```bash
docker exec kafka1 /opt/kafka/bin/kafka-consumer-groups.sh \
  --bootstrap-server kafka1:9092 \
  --describe --group orders-workers
```

This is useful to teach current offset, log-end offset, lag and consumer assignment.

## 12. Experiment: more consumers than partitions

Create 3 partitions:

```bash
go run ./cmd/admin create-topic -topic small-topic -partitions 3 -replication 3
```

Start 5 consumers with the same group. Only up to 3 consumers can own partitions at once for that topic; the remaining members have no partition work. This demonstrates why partition count bounds consumer-group parallelism.

Then increase partitions:

```bash
go run ./cmd/admin set-partitions -topic small-topic -partitions 6
```

Observe the rebalance.

## 13. Useful Make commands

```bash
make up
make ps
make create-topic TOPIC=orders PARTITIONS=6 REPLICATION=3
make consumer1
make consumer2
make consumer3
make produce TOPIC=orders COUNT=50
make describe TOPIC=orders
make groups
make down
```

Delete Kafka volumes too:

```bash
make clean
```

## 14. Quick demo helper

```bash
chmod +x scripts/demo.sh
./scripts/demo.sh
```

## Project structure

```text
golang-kafka-3node/
├── cmd/
│   ├── admin/main.go
│   ├── producer/main.go
│   └── consumer/main.go
├── internal/kafkautil/config.go
├── scripts/demo.sh
├── docker-compose.yml
├── go.mod
├── Makefile
└── README.md
```
