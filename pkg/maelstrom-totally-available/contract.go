package main

type CommitTransactionsRequest struct {
	Txn Transactions `json:"txn"`
}

type CommitTransactionsResponse struct {
	Type string       `json:"type"`
	Txn  Transactions `json:"txn"`
}

type Transaction [3]interface{}
type Transactions []Transaction
