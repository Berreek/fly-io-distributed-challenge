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

	partitionsLock := sync.RWMutex{}
	offsetsLock := sync.RWMutex{}
	partitions := make(map[string][]int)
	offsets := make(map[string]int)

	n.Handle("send", func(msg maelstrom.Message) error {
		var req SendRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling send request: %w", err)
		}

		partitionsLock.Lock()
		defer partitionsLock.Unlock()

		p, _ := partitions[req.Key]
		p = append(p, req.Msg)
		partitions[req.Key] = p

		return n.Reply(msg, &SendResponse{
			Offset: len(p) - 1,
			Type:   "send_ok",
		})
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		var req PollRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling poll request: %w", err)
		}

		msgs := make(map[string][][2]int)
		partitionsLock.Lock()
		defer partitionsLock.Unlock()

		for id, fromOffset := range req.Offsets {
			p, ok := partitions[id]
			if !ok {
				continue
			}

			var offsetsWithMessages [][2]int
			for offset, message := range p[fromOffset:] {
				offsetsWithMessages = append(offsetsWithMessages, [2]int{offset, message})
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
		var req CommitOffsetsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling commit offsets request: %w", err)
		}

		offsetsLock.Lock()
		defer offsetsLock.Unlock()

		for ID, newOffset := range req.Offsets {
			currentOffset, ok := offsets[ID]
			if !ok {
				continue
			}

			if newOffset < currentOffset {
				continue
			}

			offsets[ID] = newOffset
		}

		return n.Reply(msg, cachedCommitOffsetsResponse)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var req ListCommittedOffsetsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling list committed offsets request: %w", err)
		}

		resp := make(map[string]int)
		offsetsLock.RLock()
		defer offsetsLock.RUnlock()

		for _, key := range req.Keys {
			o, ok := offsets[key]
			if !ok {
				continue
			}
			resp[key] = o
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
