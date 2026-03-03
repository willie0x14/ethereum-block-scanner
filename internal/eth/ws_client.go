package eth

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type WSClient struct {
	client *ethclient.Client
	rpc    *rpc.Client
}

type rpcBlock struct {
	Number string `json:"number"`
	Hash   string `json:"hash"`
}

func NewWSClient(wsURL string) (*WSClient, error) {
	rpcClient, err := rpc.DialContext(context.Background(), wsURL)
	if err != nil {
		return nil, err
	}

	return &WSClient{
		client: ethclient.NewClient(rpcClient),
		rpc:    rpcClient,
	}, nil
}

func (c *WSClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return c.client.SubscribeNewHead(ctx, ch)
}

func (c *WSClient) GetBlockHash(ctx context.Context, number uint64) (string, error) {
	block, err := c.getBlockByRef(ctx, toBlockNumberArg(number))
	if err != nil {
		return "", err
	}
	if block.Hash == "" {
		return "", fmt.Errorf("empty hash for block %d", number)
	}
	return block.Hash, nil
}

func (c *WSClient) GetFinalizedHead(ctx context.Context) (uint64, error) {
	block, err := c.getBlockByRef(ctx, "finalized")
	if err != nil {
		return 0, err
	}
	return parseHexUint64(block.Number)
}

func (c *WSClient) GetSafeHead(ctx context.Context) (uint64, error) {
	block, err := c.getBlockByRef(ctx, "safe")
	if err != nil {
		return 0, err
	}
	return parseHexUint64(block.Number)
}

func (c *WSClient) getBlockByRef(ctx context.Context, ref string) (*rpcBlock, error) {
	var block *rpcBlock
	if err := c.rpc.CallContext(ctx, &block, "eth_getBlockByNumber", ref, false); err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber(%s): %w", ref, err)
	}
	if block == nil {
		return nil, fmt.Errorf("eth_getBlockByNumber(%s): nil block", ref)
	}
	return block, nil
}

func toBlockNumberArg(number uint64) string {
	if number == 0 {
		return "0x0"
	}
	return fmt.Sprintf("0x%x", number)
}

func parseHexUint64(v string) (uint64, error) {
	if v == "" {
		return 0, fmt.Errorf("empty hex value")
	}
	clean := strings.TrimPrefix(v, "0x")
	if clean == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(clean, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse hex uint64 %q: %w", v, err)
	}
	return n, nil
}

func (c *WSClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return c.client.HeaderByNumber(ctx, number)
}
