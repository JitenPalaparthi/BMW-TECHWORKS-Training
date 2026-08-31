package main

import (
	"fmt"
	"log/slog"
	"time"
)

func main() {
	go func() {
		for i := 1; i <= 10; i++ {
			slog.Info("Some log-->" + fmt.Sprint(i))
		}
	}()

	time.Sleep(time.Second * 2)
}
