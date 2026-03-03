package listener

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/willie0x14/ethereum-block-scanner/internal/repository"
	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type fakeSubscription struct {
	errCh chan error
	once  sync.Once
}

func newFakeSubscription() *fakeSubscription {
	return &fakeSubscription{errCh: make(chan error, 1)}
}

func (s *fakeSubscription) Err() <-chan error {
	return s.errCh
}

func (s *fakeSubscription) Unsubscribe() {
	s.once.Do(func() {
		close(s.errCh)
	})
}

type fakeWSClient struct {
	mu             sync.Mutex
	subscriptions  []ethereum.Subscription
	headersCh      chan<- *types.Header
	subscribeCalls int
	finalizedHead  uint64
	hashes         map[uint64]string
}

func newFakeWSClient(subs ...ethereum.Subscription) *fakeWSClient {
	return &fakeWSClient{
		subscriptions: subs,
		hashes:        make(map[uint64]string),
	}
}

func (f *fakeWSClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.headersCh = ch
	f.subscribeCalls++

	if len(f.subscriptions) == 0 {
		return nil, errors.New("no fake subscription")
	}

	idx := f.subscribeCalls - 1
	if idx >= len(f.subscriptions) {
		idx = len(f.subscriptions) - 1
	}

	return f.subscriptions[idx], nil
}

func (f *fakeWSClient) GetFinalizedHead(ctx context.Context) (uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.finalizedHead, nil
}

func (f *fakeWSClient) GetBlockHash(ctx context.Context, number uint64) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if hash, ok := f.hashes[number]; ok {
		return hash, nil
	}

	hash := fmt.Sprintf("0x%064x", number+1)
	f.hashes[number] = hash
	return hash, nil
}

func (f *fakeWSClient) pushHead(ctx context.Context, number uint64) {
	f.pushHeadWithFinalized(ctx, number, number)
}

func (f *fakeWSClient) pushHeadWithFinalized(ctx context.Context, number uint64, finalized uint64) {
	f.mu.Lock()
	ch := f.headersCh
	f.finalizedHead = finalized
	f.mu.Unlock()

	if ch != nil {
		ch <- &types.Header{Number: new(big.Int).SetUint64(number)}
	}
}

func (f *fakeWSClient) subscribeCallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.subscribeCalls
}

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout waiting for condition")
}

func TestWSListenerStart_ProcessesFinalizedBlocks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewMemoryRepository()
	svc := service.NewListenerService(repo)

	sub := newFakeSubscription()
	client := newFakeWSClient(sub)
	listener := NewWSListener(svc, client)
	listener.reconnectDelay = 10 * time.Millisecond

	done := make(chan error, 1)
	go func() {
		done <- listener.Start(ctx)
	}()

	waitUntil(t, time.Second, func() bool {
		return client.subscribeCallCount() >= 1
	})

	client.pushHead(ctx, 8)

	waitUntil(t, time.Second, func() bool {
		last, _ := repo.GetLastProcessedBlock(context.Background())
		return last == 8
	})

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("listener did not stop after context cancellation")
	}
}

func TestWSListenerStart_SkipsWhenNoFinalizedProgress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewMemoryRepository()
	svc := service.NewListenerService(repo)

	sub := newFakeSubscription()
	client := newFakeWSClient(sub)
	listener := NewWSListener(svc, client)
	listener.reconnectDelay = 10 * time.Millisecond

	done := make(chan error, 1)
	go func() {
		done <- listener.Start(ctx)
	}()

	waitUntil(t, time.Second, func() bool {
		return client.subscribeCallCount() >= 1
	})

	client.pushHeadWithFinalized(ctx, 5, 0)
	time.Sleep(50 * time.Millisecond)

	last, _ := repo.GetLastProcessedBlock(context.Background())
	assert.Equal(t, uint64(0), last)

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("listener did not stop after context cancellation")
	}
}

func TestWSListenerStart_ReconnectsAfterSubscriptionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewMemoryRepository()
	svc := service.NewListenerService(repo)

	sub1 := newFakeSubscription()
	sub2 := newFakeSubscription()
	client := newFakeWSClient(sub1, sub2)
	listener := NewWSListener(svc, client)
	listener.reconnectDelay = 10 * time.Millisecond

	done := make(chan error, 1)
	go func() {
		done <- listener.Start(ctx)
	}()

	waitUntil(t, time.Second, func() bool {
		return client.subscribeCallCount() >= 1
	})

	sub1.errCh <- errors.New("ws dropped")

	waitUntil(t, time.Second, func() bool {
		return client.subscribeCallCount() >= 2
	})

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("listener did not stop after context cancellation")
	}
}
