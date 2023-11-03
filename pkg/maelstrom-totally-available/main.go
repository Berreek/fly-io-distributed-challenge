package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewLWWKV(n)

	n.Handle("txn", func(msg maelstrom.Message) error {
		var req *CommitTransactionsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling transactions request: %w", err)
		}

		committedTransactions := make(Transactions, len(req.Txn))

		//if err := lock(kv); err != nil {
		//	return fmt.Errorf("locking: %w", err)
		//}

		for i, transaction := range req.Txn {
			key := int(transaction[1].(float64))

			if transaction[0] == "w" {
				value := int(transaction[2].(float64))
				if err := write(kv, key, value); err != nil {
					return fmt.Errorf("writing: %w", err)
				}
				committedTransactions[i] = [3]interface{}{"w", key, value}
			} else {
				value, err := read(kv, key)
				if err != nil {
					return fmt.Errorf("reading: %w", err)
				}
				committedTransactions[i] = [3]interface{}{"r", key, value}
			}
		}
		//
		//if err := unlock(kv); err != nil {
		//	return fmt.Errorf("unlocking: %w", err)
		//}

		return n.Reply(msg, &CommitTransactionsResponse{
			Type: "txn_ok",
			Txn:  committedTransactions,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func unlock(kv *maelstrom.KV) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kv.CompareAndSwap(ctx, "lock", "free", "locked", true)
	if err != nil {
		return fmt.Errorf("freeing lock in kv: %w", err)
	}

	return nil
}

func lock(kv *maelstrom.KV) error {
	maxRetries := 10
	sleepTime := 500 * time.Millisecond
	currRetries := 0

	for {
		if currRetries > maxRetries {
			return fmt.Errorf("max retries when locking key")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		err := kv.CompareAndSwap(ctx, "lock", "free", "locked", true)
		cancel()

		currRetries += 1
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}

		return nil
	}
}

func write(kv *maelstrom.KV, key, value int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := kv.Write(ctx, string(key), value); err != nil {
		return fmt.Errorf("writing to kv: %w", err)
	}

	return nil
}

func read(kv *maelstrom.KV, key int) (*int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	value, err := kv.ReadInt(ctx, string(key))

	var rpcError *maelstrom.RPCError
	if errors.As(err, &rpcError) && rpcError.Code == maelstrom.KeyDoesNotExist {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("reading from kv: %w", err)
	}

	return &value, nil
}
