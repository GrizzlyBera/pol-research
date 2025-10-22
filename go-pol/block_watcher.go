package go_pol

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"log/slog"
	"math/big"
	"time"
)

type BlockLogsWatcher struct {
	c   *ethclient.Client
	cxt context.Context

	// config
	useSubscribe bool
	batchSize    uint64
	pollInterval time.Duration

	sub ethereum.Subscription

	procLogFunc func(l types.Log) error
}

type nopSub struct{}

func (n nopSub) Unsubscribe() {
}

func (n nopSub) Err() <-chan error {
	return make(chan error)
}

var _ ethereum.Subscription = &nopSub{}

func NewDefaultWatcher(ec *ethclient.Client, cxt context.Context, procLogFunc func(l types.Log) error) *BlockLogsWatcher {
	return &BlockLogsWatcher{
		c:            ec,
		cxt:          cxt,
		procLogFunc:  procLogFunc,
		batchSize:    100,
		pollInterval: 1 * time.Second,
	}
}

func (w *BlockLogsWatcher) WatchBlockLogs(startBlock *big.Int) error {
	q := ethereum.FilterQuery{
		FromBlock: startBlock, // can be null
	}
	logsChan := make(chan types.Log, 128)

	if w.useSubscribe {
		if sub, err := w.c.SubscribeFilterLogs(w.cxt, q, logsChan); err != nil {
			return err
		} else {
			w.sub = sub
		}
	} else {
		w.sub = &nopSub{}
		// poll if subscribe is not available
		go func(cxt context.Context) {
			ticker := time.NewTicker(w.pollInterval)
			defer ticker.Stop()

			for {
				select {
				case <-cxt.Done():
					break
				case <-ticker.C:
					{
						currentBlockNum, err := w.c.BlockNumber(w.cxt)
						if err != nil {
							slog.Error("failed to filter logs - could not get blocknum", "err", err)
							continue
						}
						if currentBlockNum < q.FromBlock.Uint64() {
							continue
						} else if currentBlockNum-q.FromBlock.Uint64() > w.batchSize {
							q.ToBlock = big.NewInt(int64(q.FromBlock.Uint64() + w.batchSize - 1))
						} else {
							q.ToBlock = big.NewInt(int64(currentBlockNum))
						}

						if logs, err := w.c.FilterLogs(w.cxt, q); err != nil {
							slog.Error("failed to filter logs", "err", err)
							continue
						} else {
							maxBlock := uint64(0)
							for _, l := range logs {
								logsChan <- l
								if l.BlockNumber > maxBlock {
									maxBlock = l.BlockNumber
								}
							}
							q.FromBlock = big.NewInt(int64(maxBlock + 1))
						}
					}
				}
			}
		}(w.cxt)
	}

	for {
		select {
		// parent context cancelled
		case <-w.cxt.Done():
			return w.cxt.Err()
		// subscription error
		case err := <-w.sub.Err():
			return fmt.Errorf("error watching block logs: %v", err)
		case logEvt := <-logsChan:
			slog.Debug("log found - to be processed", slog.Any("event", logEvt))
			if err := w.procLogFunc(logEvt); err != nil {
				log.Fatalf("failed to process block log: %v", err)
			}
		}
	}
}
