package main

type AddRequest struct {
	Delta int `json:"delta"`
}

type ReadResponse struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

var (
	cachedAddResponse = map[string]string{
		"type": "add_ok",
	}
)
