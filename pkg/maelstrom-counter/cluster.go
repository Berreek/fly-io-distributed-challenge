package main

type Cluster struct {
	neighbourIDs []string
	nodeID       string
}

func NewCluster() *Cluster {
	return &Cluster{}
}

func (c *Cluster) SetNeighbourIDsAndNodeID(neighbourIDs []string, currentNodeID string) {
	var ids []string
	for _, id := range neighbourIDs {
		if id != currentNodeID {
			ids = append(ids, id)
		}
	}

	c.neighbourIDs = ids
	c.nodeID = currentNodeID
}

func (c *Cluster) NeighbourIDs() []string {
	return c.neighbourIDs
}

func (c *Cluster) NodeID() string {
	return c.nodeID
}
