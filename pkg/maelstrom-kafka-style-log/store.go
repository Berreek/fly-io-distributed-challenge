package main

import (
	"context"
	"errors"
	"fmt"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type LogsStore struct {
	kv *maelstrom.KV
}

func NewLogsStore(kv *maelstrom.KV) *LogsStore {
	return &LogsStore{kv: kv}
}

func (s *LogsStore) FetchLog(ID string) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var data []int
	err := s.kv.ReadInto(ctx, s.logsKey(ID), &data)
	var rpcError *maelstrom.RPCError
	if errors.As(err, &rpcError) && rpcError.Code == maelstrom.KeyDoesNotExist {
		return data, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetching log from kv: %w", err)
	}

	return data, nil
}

func (s *LogsStore) SaveLog(old, new []int, ID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := s.kv.CompareAndSwap(ctx, s.logsKey(ID), old, new, true)
	if err != nil {
		return fmt.Errorf("compare and swapping: %w", err)
	}

	return nil
}

func (s *LogsStore) FetchOffset(ID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	offset, err := s.kv.ReadInt(ctx, s.offsetsKey(ID))
	var rpcError *maelstrom.RPCError
	if errors.As(err, &rpcError) && rpcError.Code == maelstrom.KeyDoesNotExist {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("fetching offset from kv: %w", err)
	}

	return offset, nil
}

func (s *LogsStore) SaveOffset(ID string, currentOffset, newOffset int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	err := s.kv.CompareAndSwap(ctx, s.offsetsKey(ID), currentOffset, newOffset, true)
	if err != nil {
		return fmt.Errorf("compare and swapping: %w", err)
	}

	return nil
}

func (s *LogsStore) offsetsKey(ID string) string {
	return fmt.Sprintf("offsets-%s", ID)
}

func (s *LogsStore) logsKey(ID string) string {
	return fmt.Sprintf("logs-%s", ID)
}
