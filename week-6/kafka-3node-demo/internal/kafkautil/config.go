package kafkautil

import (
	"os"
	"strings"
)

const DefaultBrokers = "localhost:19092,localhost:29092,localhost:39092"

func Brokers() []string {
	raw := os.Getenv("KAFKA_BROKERS")
	if raw == "" {
		raw = DefaultBrokers
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
