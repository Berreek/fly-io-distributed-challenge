package main

type BroadcastRequest struct {
	Message int `json:"message"`
}

type PropagateRequest struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

type ReadResponse struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

var (
	cachedTopologyResponse = map[string]string{
		"type": "topology_ok",
	}
	cachedBroadcastResponse = map[string]string{
		"type": "broadcast_ok",
	}
	cachedPropagateResponse = map[string]string{
		"type": "propagate_ok",
	}
	cachedInitResponse = map[string]string{
		"type": "init_ok",
	}
)
