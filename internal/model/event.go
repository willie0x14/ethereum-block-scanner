package model

type Event struct {
	ID          int64  `json:"id"`
	BlockNumber int64  `json:"block_number"`
	TxHash      string `json:"tx_hash"`
	CreatedAt   int64  `json:"created_at"`
}
