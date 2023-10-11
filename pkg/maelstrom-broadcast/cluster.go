package main

type Cluster struct {
	neighbourIDs []string
}

func NewCluster() *Cluster {
	return &Cluster{}
}

func (c *Cluster) SetNeighbourIDs(neighbourIDs []string, currentNodeID string) {
	var ids []string
	for _, id := range neighbourIDs {
		if id != currentNodeID {
			ids = append(ids, id)
		}
	}

	c.neighbourIDs = ids
}

func (c *Cluster) NeighbourIDs() []string {
	return c.neighbourIDs
}
