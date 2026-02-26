package listener

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
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

func (l *WSListener) listenLoop(
	ctx context.Context,
	sub ethereum.Subscription,
	headers chan *types.Header,
) error {

	const confirmationDepth = uint64(5)
	for {
		select {

		case <-ctx.Done():
			return nil

		case err := <-sub.Err():
			return err

		case header := <-headers:
            currHead := header.Number.Uint64()
			if currHead <= confirmationDepth {
				log.Println("Waiting for enough confirmations...")
				continue
			}

			safeBlock := currHead - confirmationDepth

			last, err := l.svc.GetLastProcessedBlock(ctx)
			if err != nil {
				log.Printf("Get last processed failed: %v", err)
				continue
			}

			// 如果還沒到 safeBlock 就不用處理
			if safeBlock <= last {
				continue
			}

			log.Printf("Processing finalized blocks up to %d", safeBlock)

			// 補齊 gap + finalized
			for h := last + 1; h <= safeBlock; h++ {

				finalizedHeader, err := l.client.HeaderByNumber(
					ctx,
					big.NewInt(int64(h)),
				)
				if err != nil {
					log.Printf("Fetch header failed at %d: %v", h, err)
					break
				}

				log.Printf("Finalized block → number=%d hash=%s parent=%s",
					h,
					finalizedHeader.Hash().Hex(),
					finalizedHeader.ParentHash.Hex(),
				)

				// 直接標記 finalized
				_ = l.svc.MarkProcessed(ctx, h, finalizedHeader.Hash().Hex())
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
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("Subscribed to newHeads (WS)")

		// inner loop
		err = l.listenLoop(ctx, sub, headers)

		// 會到這邊一定是有error或者nil回傳
		sub.Unsubscribe()

		log.Printf("Subscription ended: %v", err)
		log.Println("Reconnecting in 3 seconds...")

		time.Sleep(3 * time.Second)
	}

}
