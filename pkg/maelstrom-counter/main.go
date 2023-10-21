package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	svc := NewCounterService(kv)

	n.Handle("init", func(msg maelstrom.Message) error {
		svc.InitialiseCluster(n.NodeIDs(), n.ID())

		return nil
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var req AddRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling add request: %w", err)
		}

		svc.IncreaseLocalNodeCounter(req.Delta)

		return n.Reply(msg, cachedAddResponse)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, &ReadResponse{
			Type:  "read_ok",
			Value: svc.DistributedCounter(),
		})
	})

	svc.Start()
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
