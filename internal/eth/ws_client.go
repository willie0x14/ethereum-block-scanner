package eth

import "github.com/ethereum/go-ethereum/ethclient"

func NewWSClient(wsURL string) (*ethclient.Client, error) {
	return ethclient.Dial(wsURL)
}
