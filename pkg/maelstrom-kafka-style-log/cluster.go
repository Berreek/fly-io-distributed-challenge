package main

import (
	"fmt"
	"strconv"
)

type Cluster struct {
	nodeID  string
	nodeIDs []string
}

func NewCluster() *Cluster {
	return &Cluster{}
}

func (c *Cluster) InitialiseCluster(nodeIDs []string, currentNodeID string) {
	c.nodeIDs = nodeIDs
	c.nodeID = currentNodeID
}

func (c *Cluster) IsLeaderForALog(ID string) (bool, string, error) {
	kID, err := strconv.Atoi(ID)
	if err != nil {
		return false, "", fmt.Errorf("converting log key to integer: %w", err)
	}

	leaderNodeIndex := kID % len(c.nodeIDs)
	return c.nodeIDs[leaderNodeIndex] == c.nodeID, c.nodeIDs[leaderNodeIndex], nil
}
