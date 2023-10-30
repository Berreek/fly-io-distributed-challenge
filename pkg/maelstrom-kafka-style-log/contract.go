package main

type SendRequest struct {
	Msg int    `json:"msg"`
	Key string `json:"key"`
}

type SendResponse struct {
	Offset int    `json:"offset"`
	Type   string `json:"type"`
}

type PollRequest struct {
	Offsets Offsets `json:"offsets"`
}

type PollResponse struct {
	Type string   `json:"type"`
	Msgs Messages `json:"msgs"`
}

type CommitOffsetsRequest struct {
	Offsets Offsets `json:"offsets"`
}

type ListCommittedOffsetsRequest struct {
	Keys []string `json:"keys"`
}

type ListCommittedOffsetsResponse struct {
	Type    string  `json:"type"`
	Offsets Offsets `json:"offsets"`
}

type Offsets map[string]int
type Messages map[string][][2]int

var (
	cachedCommitOffsetsResponse = map[string]string{
		"type": "commit_offsets_ok",
	}
)
