package eth

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

func NewClient(rpcURL string) *ethclient.Client {
	client, err := ethclient.DialContext(context.Background(), rpcURL)
	if err != nil {
		log.Fatalf("failed to connect to ethereum rpc: %v", err)
	}

	return client
}
