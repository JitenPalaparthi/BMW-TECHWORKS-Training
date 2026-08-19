#!/usr/bin/env bash
set -euo pipefail
TOPIC="${TOPIC:-orders}"
GROUP="${GROUP:-orders-workers}"
docker compose up -d
until docker exec kafka1 /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka1:9092 --list >/dev/null 2>&1; do sleep 2; done
go run ./cmd/admin create-topic -topic "$TOPIC" -partitions 6 -replication 3 || true
docker exec kafka1 /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka1:9092 --describe --topic "$TOPIC"
echo "Open three terminals and run:"
echo "go run ./cmd/consumer -topic $TOPIC -group $GROUP -instance consumer-1 -from-beginning"
echo "go run ./cmd/consumer -topic $TOPIC -group $GROUP -instance consumer-2 -from-beginning"
echo "go run ./cmd/consumer -topic $TOPIC -group $GROUP -instance consumer-3 -from-beginning"
echo "Then: go run ./cmd/producer -topic $TOPIC -count 30"
