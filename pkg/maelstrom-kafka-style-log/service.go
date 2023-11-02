package main

import (
	"context"
	"encoding/json"
	"fmt"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type LogsService struct {
	cluster *Cluster
	store   *LogsStore
	node    *maelstrom.Node
}

func NewLogsService() *LogsService {
	n := maelstrom.NewNode()
	return &LogsService{
		cluster: NewCluster(),
		store:   NewLogsStore(maelstrom.NewLinKV(n)),
		node:    n,
	}
}

func (s *LogsService) Node() *maelstrom.Node {
	return s.node
}

func (s *LogsService) InitialiseCluster() {
	s.cluster.InitialiseCluster(s.node.NodeIDs(), s.node.ID())
}

func (s *LogsService) SaveLogAsTheLeader(req *SendRequest) (int, error) {
	log, err := s.store.FetchLog(req.Key)
	if err != nil {
		return 0, fmt.Errorf("fetching log: %w", err)
	}

	newLog := append(log, req.Msg)
	if err = s.store.SaveLog(log, newLog, req.Key); err != nil {
		return 0, fmt.Errorf("saving log: %w", err)
	}

	return len(newLog) - 1, nil
}

func (s *LogsService) SaveLog(req *SendRequest) (int, error) {
	isLeader, leaderID, err := s.cluster.IsLeaderForALog(req.Key)
	if err != nil {
		return 0, fmt.Errorf("getting leader node ID for key %s : %w", req.Key, err)
	}

	if isLeader {
		return s.SaveLogAsTheLeader(req)
	}

	offset, err := s.sendToLeader(leaderID, req)
	if err != nil {
		return 0, fmt.Errorf("sending to leader: %w", err)
	}

	return offset, nil
}

func (s *LogsService) PollOffsets(req *PollRequest) (Messages, error) {
	msgs := make(map[string][][2]int)

	for id, fromOffset := range req.Offsets {
		l, err := s.store.FetchLog(id)
		if err != nil {
			return msgs, fmt.Errorf("fetching log with id %s : %w", id, err)
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

	return msgs, nil
}

func (s *LogsService) CommitOffsetAsLeader(req *CommitOffsetLeaderRequest) error {
	currentOffset, err := s.store.FetchOffset(req.Key)
	if err != nil {
		return fmt.Errorf("fetching offset: %w", err)
	}

	if req.Offset < currentOffset {
		return nil
	}

	if err = s.store.SaveOffset(req.Key, currentOffset, req.Offset); err != nil {
		return fmt.Errorf("saving offset: %w", err)
	}

	return nil
}

func (s *LogsService) CommitOffsets(req *CommitOffsetsRequest) error {
	for ID, newOffset := range req.Offsets {
		isLeader, leaderNodeID, err := s.cluster.IsLeaderForALog(ID)
		if err != nil {
			return fmt.Errorf("getting leader node ID for key %s : %w", ID, err)
		}

		if isLeader {
			leaderRequest := &CommitOffsetLeaderRequest{
				Offset: newOffset,
				Key:    ID,
			}
			if err := s.CommitOffsetAsLeader(leaderRequest); err != nil {
				return fmt.Errorf("committing offset as leader: %w", err)
			}
			continue
		}

		leaderRequest := &CommitOffsetLeaderRequest{
			Type:   "leader_commit_offsets",
			Offset: newOffset,
			Key:    ID,
		}
		if err = s.commitToLeader(leaderNodeID, leaderRequest); err != nil {
			return fmt.Errorf("commiting to leader: %w", err)
		}
	}

	return nil
}

func (s *LogsService) ListCommittedOffsets(req *ListCommittedOffsetsRequest) (Offsets, error) {
	offsets := make(map[string]int)
	for _, ID := range req.Keys {
		offset, err := s.store.FetchOffset(ID)
		if err != nil {
			return offsets, fmt.Errorf("fetching offset for ID %s : %w", ID, err)
		}
		offsets[ID] = offset
	}

	return offsets, nil
}

func (s *LogsService) sendToLeader(leaderID string, req *SendRequest) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	m, err := s.node.SyncRPC(ctx, leaderID, &SendLeaderRequest{
		Type: "leader_send",
		Msg:  req.Msg,
		Key:  req.Key,
	})
	if err != nil {
		return 0, fmt.Errorf("rpc call send-leader: %w", err)
	}

	var resp *SendResponse
	if err = json.Unmarshal(m.Body, &resp); err != nil {
		return 0, fmt.Errorf("unmarshalling send response: %w", err)
	}

	return resp.Offset, nil
}

func (s *LogsService) commitToLeader(leaderID string, req *CommitOffsetLeaderRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := s.node.SyncRPC(ctx, leaderID, req)
	if err != nil {
		return fmt.Errorf("rpc call commit-offset-leader: %w", err)
	}

	return nil
}
