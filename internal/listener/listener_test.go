package listener

import (
	"context"
	"errors"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/assert"
	"github.com/willie0x14/ethereum-block-scanner/internal/repository"
	"github.com/willie0x14/ethereum-block-scanner/internal/service"
)

type fakeClient struct {
	fail bool
	latest uint64
    calls int32
}

type sequenceClient struct {
	seq []uint64
	idx int32
}

func (f *fakeClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return &types.Header{
		Number: number,
	}, nil
}

func (s *sequenceClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return &types.Header{
		Number: number,
	}, nil
}


func (f *fakeClient) BlockNumber(ctx context.Context) (uint64, error) {
	// when counter, metrics, flags, state tracking, use "atomic"
    atomic.AddInt32(&f.calls, 1) // use lock-free,

	if f.fail {
		f.fail = false
		return 0, errors.New("rpc error")
	}

    return f.latest, nil
}

func (s *sequenceClient) BlockNumber(ctx context.Context) (uint64, error) {
	i := atomic.AddInt32(&s.idx, 1) - 1 // get current index

	if int(i) >= len(s.seq) {
		return s.seq[len(s.seq)-1], nil // return last one
	}

	return s.seq[i], nil
}

func TestListener_BlockNumberCalled(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

    client := &fakeClient{
		latest: 5,
	}

	repo := repository.NewMemoryRepository()
    // repo.SetLastProcessedBlock(context.Background(), 2)

    svc := service.NewListenerService(repo)
    // interval 10ms tick
    l := NewListener(svc, client, 10*time.Millisecond)

    go func() {
        time.Sleep(50 * time.Millisecond) // after 50ms cancel
        cancel()
    }()

    l.Start(ctx)

    final, _ := repo.GetLastProcessedBlock(context.Background())
    t.Log("final cursor:", final)
	assert.Equal(t, uint64(5), final)
}


func TestListener_NoNewBlock(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &fakeClient{
		latest: 5,
	}

	repo := repository.NewMemoryRepository()
	// repo.SetLastProcessedBlock(context.Background(), 5)

	svc := service.NewListenerService(repo)
	l := NewListener(svc, client, 10*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	l.Start(ctx)

	final, _ := repo.GetLastProcessedBlock(context.Background())

	t.Log("final cursor:", final)

	assert.Equal(t, uint64(5), final)
}


func TestListener_RetryOnRPCError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &fakeClient{
		latest: 10,
		fail: true,
	}

	repo := repository.NewMemoryRepository()

	svc := service.NewListenerService(repo)
	l := NewListener(svc, client, 10*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	l.Start(ctx)

	calls := atomic.LoadInt32(&client.calls)

	t.Log("rpc calls:", calls)

	assert.Greater(t, calls, int32(1))
}


func TestListener_SequenceBlockNumber(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &sequenceClient{
		seq: []uint64{10, 9, 10},
		idx: 0,
	}

	repo := repository.NewMemoryRepository()
	svc := service.NewListenerService(repo) // DI
	l := NewListener(svc, client, 10*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	l.Start(ctx)

	final, _ := repo.GetLastProcessedBlock(context.Background())
	assert.Equal(t, uint64(10), final) // assert.Equle(t, expected, actual)

}
