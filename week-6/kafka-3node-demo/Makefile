.PHONY: up down clean ps create-topic add-partitions produce consumer1 consumer2 consumer3 describe groups fmt
TOPIC ?= orders
PARTITIONS ?= 6
REPLICATION ?= 3
GROUP ?= orders-workers
COUNT ?= 20
up:
	docker compose up -d
down:
	docker compose down
clean:
	docker compose down -v
ps:
	docker compose ps
create-topic:
	go run ./cmd/admin create-topic -topic $(TOPIC) -partitions $(PARTITIONS) -replication $(REPLICATION)
add-partitions:
	go run ./cmd/admin add-partitions -topic $(TOPIC) -add 3
produce:
	go run ./cmd/producer -topic $(TOPIC) -count $(COUNT)
consumer1:
	go run ./cmd/consumer -topic $(TOPIC) -group $(GROUP) -instance consumer-1 -from-beginning
consumer2:
	go run ./cmd/consumer -topic $(TOPIC) -group $(GROUP) -instance consumer-2 -from-beginning
consumer3:
	go run ./cmd/consumer -topic $(TOPIC) -group $(GROUP) -instance consumer-3 -from-beginning
describe:
	docker exec kafka1 /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka1:9092 --describe --topic $(TOPIC)
groups:
	docker exec kafka1 /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server kafka1:9092 --list
fmt:
	gofmt -w ./cmd ./internal
