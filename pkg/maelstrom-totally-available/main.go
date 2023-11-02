package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	store := make(map[int]int)
	var lock sync.Mutex

	n.Handle("txn", func(msg maelstrom.Message) error {
		var req *CommitTransactionsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling transactions request: %w", err)
		}

		committedTransactions := make(Transactions, len(req.Txn))

		lock.Lock()
		defer lock.Unlock()

		for i, transaction := range req.Txn {
			key := int(transaction[1].(float64))

			if transaction[0] == "w" {
				value := int(transaction[2].(float64))
				store[key] = value
				committedTransactions[i] = [3]interface{}{"w", key, value}
			} else {
				value, ok := store[key]
				if !ok {
					committedTransactions[i] = [3]interface{}{"r", key, nil}
				} else {
					committedTransactions[i] = [3]interface{}{"r", key, value}
				}
			}
		}

		return n.Reply(msg, &CommitTransactionsResponse{
			Type: "txn_ok",
			Txn:  committedTransactions,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
