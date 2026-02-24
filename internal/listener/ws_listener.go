package listener

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type WSListener struct {
	svc    *service.ListenerService
	client *ethclient.Client
}

func NewWSListener(svc *service.ListenerService, client *ethclient.Client) *WSListener {
	return &WSListener{
		svc:    svc,
		client: client,
	}
}

func (l *WSListener) Start(ctx context.Context) error {
	headers := make(chan *types.Header)

	sub, err := l.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return err
	}

	log.Println("subscribed to newHeads (WS)")

	for {
		select {
		case <-ctx.Done():
			log.Println("WS listener context cancelled")
			return nil

		case err := <-sub.Err():
			// WS 斷線、節點問題、訂閱失效等都會來這
			return err

		case header := <-headers:
			// 先做到 MVP：收到新區塊就印出來
			log.Printf("new block → number=%d hash=%s parent=%s",
				header.Number.Uint64(),
				header.Hash().Hex(),
				header.ParentHash.Hex(),
			)

			// 你後面 B 階段要做的事：丟給 service 處理
			// 這裡我先用「最保守」的方式：只把新高度交給 service
			// 你可以在 svc 裡面做去重、確認數、reorg 等
			l.svc.OnNewHead(ctx, header)
		}
	}
}
