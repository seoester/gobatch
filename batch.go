package gobatch

import (
	"context"
	"sync"
	"time"

	"github.com/MasterOfBinary/gobatch/processor"
	"github.com/MasterOfBinary/gobatch/source"
)

type batchImpl struct {
	minTime         time.Duration
	minItems        uint64
	maxTime         time.Duration
	maxItems        uint64
	readConcurrency uint64

	running        bool

	src  source.Source
	proc processor.Processor

	items chan interface{}
	errs  chan error
	done  chan struct{}

	setupOnce sync.Once
	mu        sync.Mutex
}

func (b *batchImpl) Go(ctx context.Context, s source.Source, p processor.Processor) <-chan error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		b.errs <- ErrConcurrentGoCalls
		return b.errs
	}

	b.running = true
	b.items = make(chan interface{})
	b.errs = make(chan error)
	b.done = make(chan struct{})
	b.src = s
	b.proc = p

	go b.doReaders(ctx)
	go b.doProcessors(ctx)

	return b.errs
}

func (b *batchImpl) Done() <-chan struct{} {
    b.mu.Lock()
	defer b.mu.Unlock()
    return b.done
}

func (b *batchImpl) doReaders(ctx context.Context) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	
    if readConcurrency > 0 {
		var wg sync.WaitGroup
		for i := 0; i < b.readConcurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				read(ctx)
			}()
		}
		wg.Wait()
	} else {
		err := errors.New("Read concurrency is 0")
		b.errs <- err
	}

	b.mu.Lock()
	close(b.items)
	close(b.errs)
	b.mu.Unlock()
}

func (b *batchImpl) doProcessors(ctx context.Context) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	
	// ...

	// Once processors are complete, everything is
	b.mu.Lock()
	b.running = false
	close(b.done)
	b.mu.Unlock()
}

func (b *batchImpl) read(ctx context.Context) {
	items := make(chan interface{})
	errs := make(chan error)

	go b.src.Read(ctx, items, errs)

	// Read should close the channels when the context is done, so we don't check
	// ctx.Done() here. Otherwise we might return before Read is completely
	// finished. The way we know we've received everything from Read is
	// when the channels have been closed.
	var itemsClosed, errsClosed bool
	for {
		select {
		case item, ok := <-items:
			if ok {
				b.items <- item
			} else {
				itemsClosed = true
			}
		case err, ok := <-errs:
			if ok {
			    wrappedErr := newSourceError(err)
				b.errs <- wrappedErr
			} else {
				errsClosed = true
			}
		}
		if itemsClosed && errsClosed {
			break
		}
	}
}