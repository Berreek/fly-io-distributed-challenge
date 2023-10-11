package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	svc := NewBroadcastingService()
	node := svc.Node()

	node.Handle("propagate", func(msg maelstrom.Message) error {
		var req *PropagateRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling propagate request: %w", err)
		}

		svc.SaveMessagesAsReplica(req.Messages)
		return node.Reply(msg, cachedPropagateResponse)
	})

	node.Handle("broadcast", func(msg maelstrom.Message) error {
		var req *BroadcastRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling broadcast request: %w", err)
		}

		svc.SaveMessageAsLeader(req.Message, msg)
		return node.Reply(msg, cachedBroadcastResponse)
	})

	node.Handle("read", func(msg maelstrom.Message) error {
		resp := &ReadResponse{
			Type:     "read_ok",
			Messages: svc.ReadMessages(),
		}

		return node.Reply(msg, resp)
	})

	node.Handle("topology", func(msg maelstrom.Message) error {
		svc.InitialiseCluster(node.NodeIDs(), node.ID())
		return node.Reply(msg, cachedTopologyResponse)
	})

	svc.Start()
	if err := node.Run(); err != nil {
		log.Fatal(err)
	}

	svc.Stop()
}
