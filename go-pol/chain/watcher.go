package chain

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/blocktracker"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/tracker"
	boltdbStore "github.com/umbracle/ethgo/tracker/store/boltdb"
	"os"
	"os/signal"
	"path"
	"syscall"
)

// simple watcher that watches one event on one contract
type ContractEventWatcher struct {
	addr     ethgo.Address
	evtIds   map[ethgo.Hash]bool
	chainUrl string
	dataDir  string

	cancel context.CancelFunc
}

func NewWatcher(addr ethgo.Address, evts []*abi.Event, chainUrl, dataDir string) *ContractEventWatcher {

	mapEvts := make(map[ethgo.Hash]bool)
	for _, evt := range evts {
		mapEvts[evt.ID()] = true
	}

	return &ContractEventWatcher{
		addr:     addr,
		evtIds:   mapEvts,
		chainUrl: chainUrl,
		dataDir:  dataDir,
	}
}

func (w *ContractEventWatcher) GetCancelFunc() context.CancelFunc {
	return w.cancel
}

func (w *ContractEventWatcher) Start(startBlock uint64, handleAdded func(log *ethgo.Log) error) error {
	ec, err := jsonrpc.NewClient(w.chainUrl)
	if err != nil {
		return err
	}

	dbPath := path.Join(w.dataDir, fmt.Sprintf("%v.db", w.addr.String()[2:10]))
	store, err := boltdbStore.New(dbPath)
	if err != nil {
		return err
	}

	bt := blocktracker.NewBlockTracker(ec.Eth(), blocktracker.WithBlockMaxBacklog(100))

	tt, err := tracker.NewTracker(ec.Eth(),
		tracker.WithBatchSize(2000),
		tracker.WithStore(store),
		tracker.WithBlockTracker(bt),
		// tracker.WithEtherscan(os.Getenv("ETHERSCAN_APIKEY")),
		tracker.WithFilter(&tracker.FilterConfig{
			Async: true,
			Address: []ethgo.Address{
				w.addr,
			},
			Start: startBlock,
		}),
	)
	if err != nil {
		return err
	}

	lastBlock, err := tt.GetLastBlock()
	if err != nil {
		return err
	}
	if lastBlock != nil {
		log.Infof("watcher %v found last block %v", dbPath, lastBlock.Number)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	go func() {
		go func() {
			if err := tt.Sync(ctx); err != nil {
				log.Errorf("ERROR in watcher %v: %v", dbPath, err.Error())
			}
		}()

		go func() {
			cntBlocks := 0
			for {
				select {
				case evnt := <-tt.EventCh:
					newLastBlock, _ := tt.GetLastBlock()
					cntBlocks++
					if len(evnt.Added) > 0 {
						log.Infof("got watcher events: %v, last block %v, %v cnt added", dbPath, newLastBlock.Number, len(evnt.Added))
					} else if cntBlocks%100 == 0 {
						log.Infof("watcher processing events: %v, last block %v, cnt blocks %v", dbPath, newLastBlock.Number, cntBlocks)
					}
					for _, evtLog := range evnt.Added {
						if w.evtIds[evtLog.Topics[0]] {
							if err = handleAdded(evtLog); err != nil {
								log.Errorf("ERROR handling event in watcher %v: %v", dbPath, err.Error())
							}
						}
					}
				case <-tt.DoneCh:
					fmt.Println("historical sync done")
				}
			}
		}()

	}()

	w.cancel = cancelFn
	return nil

}

func handleSignals(cancelFn context.CancelFunc) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	<-signalCh

	gracefulCh := make(chan struct{})
	go func() {
		cancelFn()
		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return 1
	case <-gracefulCh:
		return 0
	}
}
