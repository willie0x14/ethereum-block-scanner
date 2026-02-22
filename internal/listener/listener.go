package listener

import (
	"context"
	"log"
	"time"

	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type BlockClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
}

type Listener struct {
	svc *service.ListenerService
	client BlockClient
	tickerInterval time.Duration
}

func NewListener(svc *service.ListenerService, client BlockClient, interval time.Duration) *Listener {
	return &Listener{
		svc:    svc,
		client: client,
		tickerInterval: interval,
	}
}

func (l *Listener) Start(ctx context.Context) {
	var lastSeen uint64

	log.Println("Listener started...")

    ticker := time.NewTicker(l.tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Listener stopped")
			return

		case <-ticker.C:
			blockNumber, err := l.client.BlockNumber(ctx)
			if err != nil {
				log.Println("Failed to get block number:", err)
				continue
			}

			if blockNumber != lastSeen {
				log.Println("current block:", blockNumber)
				lastSeen = blockNumber
			}
		}
	}
}
