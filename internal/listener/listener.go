package listener

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type Listener struct {
	svc *service.ListenerService
}

func NewListener(svc *service.ListenerService) *Listener {
	return &Listener{
		svc: svc,
	}
}

func (l *Listener) Start(ctx context.Context) {
	var lastSeen uint64

	log.Println("Listener started...")

    // 1) 建立 RPC client（只建一次）
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		log.Fatal("ETH_RPC_URL is not set")
	}

	client, err := ethclient.DialContext(ctx, rpcURL) // 建立連線
	if err != nil {
		log.Fatalf("failed to connect to ethereum rpc: %v", err)
	}
	defer client.Close()


	// 每 3 秒送一次訊號到 ticker.C
	// C：channel（type is <-chan time.Time）
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Listener stopped")
			return
		case <-ticker.C:
			blockNumber, err := client.BlockNumber(ctx)
			if err != nil {
				log.Println("Failed to get block number:", err)
				continue
			}

			if blockNumber != lastSeen {
                log.Println("current block:", blockNumber)
				lastSeen = blockNumber
			}


			// 之後這裡可以呼叫 l.svc.ProcessBlock(...)
		}
	}
}
