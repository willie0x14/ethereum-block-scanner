package listener

import (
    "context"
    "sync/atomic"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

type fakeClient struct {
    calls int32
}

func (f *fakeClient) BlockNumber(ctx context.Context) (uint64, error) {
	// when counter, metrics, flags, state tracking, use "atomic"
    atomic.AddInt32(&f.calls, 1) // use lock-free,
    return 123, nil
}

func TestListener_BlockNumberCalled(t *testing.T) {
    client := &fakeClient{}

    // interval 10ms tick
    l := NewListener(nil, client, 10*time.Millisecond)

    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        time.Sleep(50 * time.Millisecond) // after 50ms cancel
        cancel()
    }()

    l.Start(ctx)

    // assert calls > 0
    assert.Greater(t, atomic.LoadInt32(&client.calls), int32(0))
}
