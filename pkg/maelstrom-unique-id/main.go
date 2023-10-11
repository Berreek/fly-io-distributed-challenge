package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Response struct {
	Id   int64  `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

func main() {
	n := maelstrom.NewNode()
	n.Handle("generate", func(msg maelstrom.Message) error {
		nodeID, err := strconv.Atoi(msg.Dest[1:])
		if err != nil {
			return fmt.Errorf("parsing node ID: %w", err)
		}

		clientID, err := strconv.Atoi(msg.Src[1:])
		if err != nil {
			return fmt.Errorf("parsing client ID: %w", err)
		}

		now := time.Now().UnixMilli()
		id := rand.NewSource(now + int64(nodeID) + int64(clientID) + rand.Int63()).Int63()
		resp := &Response{Id: id, Type: "generate_ok"}

		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
