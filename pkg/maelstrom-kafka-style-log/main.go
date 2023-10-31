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
	kv := maelstrom.NewLinKV(n)

	n.Handle("send", func(msg maelstrom.Message) error {
		var req *SendRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling send request: %w", err)
		}

		log, err := fetchLog(kv, req.Key)
		if err != nil {
			return fmt.Errorf("fetching log: %w", err)
		}

		updated, err := appendDataAndSaveLog(kv, log, req)
		if err != nil {
			return fmt.Errorf("updating log: %w", err)
		}

		return n.Reply(msg, &SendResponse{
			Offset: len(updated) - 1,
			Type:   "send_ok",
		})
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		var req *PollRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling poll request: %w", err)
		}

		msgs := make(map[string][][2]int)

		for id, fromOffset := range req.Offsets {
			l, err := fetchLog(kv, id)
			if err != nil {
				return fmt.Errorf("fetching log with id %s : %w", id, err)
			}

			var offsetsWithMessages [][2]int
			for i, message := range l[fromOffset:] {
				offsetsWithMessages = append(offsetsWithMessages, [2]int{i + fromOffset, message})
			}
			if len(offsetsWithMessages) == 0 {
				continue
			}

			msgs[id] = offsetsWithMessages
		}

		return n.Reply(msg, &PollResponse{
			Type: "poll_ok",
			Msgs: msgs,
		})
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		var req *CommitOffsetsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling commit offsets request: %w", err)
		}

		for ID, newOffset := range req.Offsets {
			offset, err := fetchOffset(kv, ID)
			if err != nil {
				return fmt.Errorf("fetching offset: %w", err)
			}

			if newOffset < offset {
				continue
			}

			if err = saveOffset(kv, ID, offset, newOffset); err != nil {
				return fmt.Errorf("saving offset: %w", err)
			}
		}

		return n.Reply(msg, cachedCommitOffsetsResponse)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var req *ListCommittedOffsetsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling list committed offsets request: %w", err)
		}

		resp := make(map[string]int)

		for _, key := range req.Keys {
			offset, err := fetchOffset(kv, key)
			if err != nil {
				return fmt.Errorf("fetching offset for ID %s : %w", key, err)
			}
			resp[key] = offset
		}

		return n.Reply(msg, &ListCommittedOffsetsResponse{
			Offsets: resp,
			Type:    "list_committed_offsets_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func fetchLog(kv *maelstrom.KV, ID string) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var data []int
	err := kv.ReadInto(ctx, fmt.Sprintf("logs-%s", ID), &data)
	var rpcError *maelstrom.RPCError
	if errors.As(err, &rpcError) && rpcError.Code == maelstrom.KeyDoesNotExist {
		return data, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching log from kv: %w", err)
	}

	return data, nil
}

func fetchOffset(kv *maelstrom.KV, ID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	offset, err := kv.ReadInt(ctx, fmt.Sprintf("offsets-%s", ID))
	var rpcError *maelstrom.RPCError
	if errors.As(err, &rpcError) && rpcError.Code == maelstrom.KeyDoesNotExist {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("fetching log from kv: %w", err)
	}

	return offset, nil
}

func appendDataAndSaveLog(kv *maelstrom.KV, log []int, req *SendRequest) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	updated := append(log, req.Msg)
	err := kv.CompareAndSwap(ctx, fmt.Sprintf("logs-%s", req.Key), log, updated, true)
	if err != nil {
		return nil, fmt.Errorf("compare and swapping: %w", err)
	}

	return updated, nil
}

func saveOffset(kv *maelstrom.KV, ID string, currentOffset, newOffset int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := kv.CompareAndSwap(ctx, fmt.Sprintf("offsets-%s", ID), currentOffset, newOffset, true)
	if err != nil {
		return fmt.Errorf("compare and swapping: %w", err)
	}

	return nil
}
