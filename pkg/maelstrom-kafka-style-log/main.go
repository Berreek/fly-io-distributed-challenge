package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var defaultTimeout = 2 * time.Second

func main() {

	svc := NewLogsService()
	n := svc.Node()

	n.Handle("init", func(msg maelstrom.Message) error {
		svc.InitialiseCluster()
		return nil
	})

	n.Handle("leader_send", func(msg maelstrom.Message) error {
		var req *SendRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling send request: %w", err)
		}

		offset, err := svc.SaveLogAsTheLeader(req)
		if err != nil {
			return fmt.Errorf("saving log as the leader: %w", err)
		}

		return n.Reply(msg, &SendResponse{
			Offset: offset,
			Type:   "leader_send_ok",
		})
	})

	n.Handle("send", func(msg maelstrom.Message) error {
		var req *SendRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling send request: %w", err)
		}

		offset, err := svc.SaveLog(req)
		if err != nil {
			return fmt.Errorf("saving log: %w", err)
		}

		return n.Reply(msg, &SendResponse{
			Offset: offset,
			Type:   "send_ok",
		})
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		var req *PollRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling poll request: %w", err)
		}

		msgs, err := svc.PollOffsets(req)
		if err != nil {
			return fmt.Errorf("polling offsets: %w", err)
		}

		return n.Reply(msg, &PollResponse{
			Type: "poll_ok",
			Msgs: msgs,
		})
	})

	n.Handle("leader_commit_offsets", func(msg maelstrom.Message) error {
		var req *CommitOffsetLeaderRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling commit offset leader request: %w", err)
		}

		if err := svc.CommitOffsetAsLeader(req); err != nil {
			return fmt.Errorf("commiting offset as leader: %w", err)
		}

		return n.Reply(msg, cachedCommitOffsetsLeaderResponse)
	})

	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		var req *CommitOffsetsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling commit offsets request: %w", err)
		}

		if err := svc.CommitOffsets(req); err != nil {
			return fmt.Errorf("committing offsets: %w", err)
		}

		return n.Reply(msg, cachedCommitOffsetsResponse)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var req *ListCommittedOffsetsRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return fmt.Errorf("unmarshaling list committed offsets request: %w", err)
		}

		offsets, err := svc.ListCommittedOffsets(req)
		if err != nil {
			return fmt.Errorf("list committed offsets: %w", err)
		}

		return n.Reply(msg, &ListCommittedOffsetsResponse{
			Offsets: offsets,
			Type:    "list_committed_offsets_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
