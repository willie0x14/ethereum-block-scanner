package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

    "github.com/willie0x14/ethereum-block-scanner/internal/eth"
	"github.com/willie0x14/ethereum-block-scanner/internal/api"
	"github.com/willie0x14/ethereum-block-scanner/internal/listener"
	"github.com/willie0x14/ethereum-block-scanner/internal/repository"
	"github.com/willie0x14/ethereum-block-scanner/internal/service"
	"github.com/willie0x14/ethereum-block-scanner/internal/config"

)


func main() {

	cfg := config.Load()

	// load .env
	err := godotenv.Load()
    if err != nil {
        log.Println(".env not found, using system env")
    }

	// root context
	ctx, cancel := context.WithCancel(context.Background()) // 可以呼叫cancel通知所有ctx的goroutine結束
	defer cancel()

	var wg sync.WaitGroup // 等待啟動的goroutine做完再退出

	// repository
	repo := repository.NewMemoryRepository()

	// service
	svc := service.NewListenerService(repo)

	// eth client
	rpcURL := os.Getenv("ETH_RPC_URL")
    ethClient := eth.NewClient(rpcURL)

	// initialize cursor to latest block to avoid backfilling from block 1 on startup
	latestBlock, err := ethClient.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("failed to get latest block number: %v", err)
	}
	repo.SetLastProcessedBlock(ctx, latestBlock)
	log.Printf("Initialized last processed block to latest: %d", latestBlock)

	// listener
	blockListener := listener.NewListener(svc, ethClient, cfg.PollInterval)

	wg.Add(1) // 多一個goroutine需要等待
	go func() {
		defer wg.Done()
		blockListener.Start(ctx) // start listener loop
	}()

	// ===== Gin router =====
	gin.SetMode(gin.ReleaseMode)

	handler := api.NewHandler(svc)
	router := handler.Router() // Router() 要回傳 *gin.Engine

	// ===== http server (包 Gin 以便 graceful shutdown) =====
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("HTTP server started at :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed { // 阻塞式
			log.Fatal(err)
		}
 	}()

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // 卡住, 直到收到signal
	log.Println("Shutting down...")

	cancel() // cancel all  goroutines

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Println("Server shutdown error:", err)
	}

	wg.Wait()
	log.Println("Server exited cleanly")

}
