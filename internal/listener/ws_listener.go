package listener

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type WSBlockClient interface {
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	GetFinalizedHead(ctx context.Context) (uint64, error)
	GetBlockHash(ctx context.Context, number uint64) (string, error)
}

type WSListener struct {
	svc            *service.ListenerService
	client         WSBlockClient
	reconnectDelay time.Duration
}

func NewWSListener(svc *service.ListenerService, client WSBlockClient) *WSListener {
	return &WSListener{
		svc:            svc,
		client:         client,
		reconnectDelay: 3 * time.Second,
	}
}

func (l *WSListener) listenLoop(
	ctx context.Context,
	sub ethereum.Subscription,
	headers chan *types.Header,
) error {

	for {
		select {

		case <-ctx.Done():
			return nil

		case err := <-sub.Err():
			return err

		case <-headers:
			if err := l.svc.SyncFinalized(ctx, l.client.GetFinalizedHead, l.client.GetBlockHash); err != nil {
				log.Printf("Sync finalized failed: %v", err)
			}
		}
	}
}

func (l *WSListener) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("WS listener shutting down")
			return nil
		default: // doNothing
		}

		log.Println("Connecting to WS...")

		headers := make(chan *types.Header)

		sub, err := l.client.SubscribeNewHead(ctx, headers)
		if err != nil {
			log.Printf("Subscribe failed: %v", err)
			if ctx.Err() != nil {
				return nil
			}
			time.Sleep(l.reconnectDelay)
			continue
		}

		log.Println("Subscribed to newHeads (WS)")

		// inner loop
		err = l.listenLoop(ctx, sub, headers)

		// 會到這邊一定是有error或者nil回傳
		sub.Unsubscribe()

		if ctx.Err() != nil {
			log.Println("WS listener shutting down")
			return nil
		}

		log.Printf("Subscription ended: %v", err)
		log.Printf("Reconnecting in %s...", l.reconnectDelay)

		time.Sleep(l.reconnectDelay)
	}

}
