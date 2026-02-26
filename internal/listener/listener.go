package listener

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type BlockClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)

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
	log.Println("Listener started...")

    ticker := time.NewTicker(l.tickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Listener stopped")
			return

		case <-ticker.C:
			latest, err := l.client.BlockNumber(ctx)
			if err != nil {
				log.Println("Failed to get block number:", err)
				continue
			}

			cursor, err:= l.svc.GetLastProcessedBlock(ctx)
			if err != nil {
				log.Println("Failed to get last processed block:", err)
				continue
			}

			for b := cursor + 1; b <= latest; b++ {
				hdr, err := l.client.HeaderByNumber(ctx, big.NewInt(int64(b)))
				if err != nil {
					log.Println("failed to get header:", err)
					break
				}
				l.svc.ProcessBlock(ctx, b, hdr.Hash().Hex())
			}


		}
	}
}
