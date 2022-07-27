package complete

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/semaphore"

	"github.com/onflow/flow-go/ledger"
	"github.com/onflow/flow-go/ledger/complete/mtrie/trie"
	realWAL "github.com/onflow/flow-go/ledger/complete/wal"
	"github.com/onflow/flow-go/module/lifecycle"
	"github.com/onflow/flow-go/module/observable"
)

// WALTrieUpdate is a message communicated through channel between Ledger and Compactor.
type WALTrieUpdate struct {
	Update   *ledger.TrieUpdate // Update data needs to be encoded and saved in WAL.
	ResultCh chan<- error       // ResultCh channel is used to send WAL update result from Compactor to Ledger.
	TrieCh   <-chan *trie.MTrie // TrieCh channel is used to send new trie from Ledger to Compactor.
}

// checkpointResult is a message to communicate checkpointing number and error if any.
type checkpointResult struct {
	num int
	err error
}

// Compactor is a long-running goroutine responsible for:
// - writing WAL record from trie update,
// - starting checkpointing async when enough segments are finalized.
//
// Compactor communicates with Ledger through channels
// to ensure that by the end of any trie update processing,
// update is written to WAL and new trie is pushed to trie queue.
//
// Compactor stores pointers to tries in ledger state in a fix-sized
// checkpointing queue (FIFO).  Checkpointing queue is decoupled from
// main ledger state to allow separate optimizaiton, etc.
// NOTE: ledger state and checkpointing queue may contain different tries.
type Compactor struct {
	checkpointer       *realWAL.Checkpointer
	wal                realWAL.LedgerWAL
	trieQueue          *realWAL.TrieQueue
	logger             zerolog.Logger
	lm                 *lifecycle.LifecycleManager
	observers          map[observable.Observer]struct{}
	checkpointDistance uint
	checkpointsToKeep  uint
	stopCh             chan chan struct{}
	trieUpdateCh       <-chan *WALTrieUpdate
}

func NewCompactor(
	l *Ledger,
	w realWAL.LedgerWAL,
	logger zerolog.Logger,
	checkpointCapacity uint,
	checkpointDistance uint,
	checkpointsToKeep uint,
) (*Compactor, error) {
	if checkpointDistance < 1 {
		checkpointDistance = 1
	}

	checkpointer, err := w.NewCheckpointer()
	if err != nil {
		return nil, err
	}

	trieUpdateCh := l.TrieUpdateChan()
	if trieUpdateCh == nil {
		return nil, errors.New("failed to get valid trie update channel from ledger")
	}

	tries, err := l.Tries()
	if err != nil {
		return nil, err
	}

	trieQueue := realWAL.NewTrieQueueWithValues(checkpointCapacity, tries)

	return &Compactor{
		checkpointer:       checkpointer,
		wal:                w,
		trieQueue:          trieQueue,
		logger:             logger,
		stopCh:             make(chan chan struct{}),
		trieUpdateCh:       trieUpdateCh,
		observers:          make(map[observable.Observer]struct{}),
		lm:                 lifecycle.NewLifecycleManager(),
		checkpointDistance: checkpointDistance,
		checkpointsToKeep:  checkpointsToKeep,
	}, nil
}

func (c *Compactor) Subscribe(observer observable.Observer) {
	var void struct{}
	c.observers[observer] = void
}

func (c *Compactor) Unsubscribe(observer observable.Observer) {
	delete(c.observers, observer)
}

// Ready periodically fires Run function, every `interval`
func (c *Compactor) Ready() <-chan struct{} {
	c.lm.OnStart(func() {
		go c.run()
	})
	return c.lm.Started()
}

func (c *Compactor) Done() <-chan struct{} {
	c.lm.OnStop(func() {
		// Signal Compactor goroutine to stop
		doneCh := make(chan struct{})
		c.stopCh <- doneCh

		// Wait for Compactor goroutine to stop
		<-doneCh

		// Shut down WAL component.
		<-c.wal.Done()

		// Notify observers
		for observer := range c.observers {
			observer.OnComplete()
		}
	})
	return c.lm.Stopped()
}

func (c *Compactor) run() {

	// checkpointSem is used to limit checkpointing to one.
	// If previous checkpointing isn't finished when enough segments
	// are finalized for next checkpointing, retry checkpointing
	// again when next segment is finalized.
	// This avoids having more tries in memory than needed.
	checkpointSem := semaphore.NewWeighted(1)

	checkpointResultCh := make(chan checkpointResult, 1)

	// Get active segment number.
	// activeSegmentNum is updated when record is written to a new segment.
	_, activeSegmentNum, err := c.wal.Segments()
	if err != nil {
		c.logger.Error().Err(err).Msg("compactor failed to get active segment number")
		activeSegmentNum = -1
	}

	lastCheckpointNum, err := c.checkpointer.LatestCheckpoint()
	if err != nil {
		c.logger.Error().Err(err).Msg("compactor failed to get last checkpoint number")
		lastCheckpointNum = -1
	}

	// Compute next checkpoint number.
	// nextCheckpointNum is updated when
	// - checkpointing starts, fails to start, or fails.
	// - tries snapshot fails.
	// NOTE: next checkpoint number must >= active segment num.
	// We can't reuse mtrie state to checkpoint tries in older segments.
	nextCheckpointNum := lastCheckpointNum + int(c.checkpointDistance)
	if activeSegmentNum > nextCheckpointNum {
		nextCheckpointNum = activeSegmentNum
	}

	ctx, cancel := context.WithCancel(context.Background())

Loop:
	for {
		select {

		case doneCh := <-c.stopCh:
			defer close(doneCh)
			cancel()
			break Loop

		case checkpointResult := <-checkpointResultCh:
			if checkpointResult.err != nil {
				c.logger.Error().Err(checkpointResult.err).Msgf(
					"compactor failed to checkpoint %d", checkpointResult.num,
				)

				// Retry checkpointing after active segment is finalized.
				nextCheckpointNum = activeSegmentNum
			}

		case update, ok := <-c.trieUpdateCh:
			if !ok {
				// trieUpdateCh channel is closed.
				// Wait for stop signal from c.stopCh
				continue
			}

			var checkpointNum int
			var checkpointTries []*trie.MTrie
			activeSegmentNum, checkpointNum, checkpointTries =
				c.processTrieUpdate(update, c.trieQueue, activeSegmentNum, nextCheckpointNum)

			if checkpointTries == nil {
				// Not enough segments for checkpointing (nextCheckpointNum >= activeSegmentNum)
				continue
			}

			// Try to checkpoint
			if checkpointSem.TryAcquire(1) {

				// Compute next checkpoint number
				nextCheckpointNum = checkpointNum + int(c.checkpointDistance)

				go func() {
					defer checkpointSem.Release(1)
					err := c.checkpoint(ctx, checkpointTries, checkpointNum)
					checkpointResultCh <- checkpointResult{checkpointNum, err}
				}()
			} else {
				// Failed to get semaphore because checkpointing is running.
				// Try again when active segment is finalized.
				c.logger.Info().Msgf("compactor delayed checkpoint %d because prior checkpointing is ongoing", nextCheckpointNum)
				nextCheckpointNum = activeSegmentNum
			}
		}
	}

	// Drain and process remaining trie updates in channel.
	for update := range c.trieUpdateCh {
		_, _, err := c.wal.RecordUpdate(update.Update)
		select {
		case update.ResultCh <- err:
		default:
		}
	}

	// Don't wait for checkpointing to finish because it might take too long.
}

func (c *Compactor) checkpoint(ctx context.Context, tries []*trie.MTrie, checkpointNum int) error {

	err := createCheckpoint(c.checkpointer, c.logger, tries, checkpointNum)
	if err != nil {
		return fmt.Errorf("cannot create checkpoints: %w", err)
	}

	// Return if context is canceled.
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	err = cleanupCheckpoints(c.checkpointer, int(c.checkpointsToKeep))
	if err != nil {
		return fmt.Errorf("cannot cleanup checkpoints: %w", err)
	}

	if checkpointNum > 0 {
		for observer := range c.observers {
			// Don't notify observer if context is canceled.
			// observer.OnComplete() is called when Compactor starts shutting down,
			// which may close channel that observer.OnNext() uses to send data.
			select {
			case <-ctx.Done():
				return nil
			default:
				observer.OnNext(checkpointNum)
			}
		}
	}

	return nil
}

func createCheckpoint(checkpointer *realWAL.Checkpointer, logger zerolog.Logger, tries []*trie.MTrie, checkpointNum int) error {

	logger.Info().Msgf("serializing checkpoint %d", checkpointNum)

	startTime := time.Now()

	writer, err := checkpointer.CheckpointWriter(checkpointNum)
	if err != nil {
		return fmt.Errorf("cannot generate checkpoint writer: %w", err)
	}
	defer func() {
		closeErr := writer.Close()
		// Return close error if there isn't any prior error to return.
		if err == nil {
			err = closeErr
		}
	}()

	err = realWAL.StoreCheckpoint(writer, tries...)
	if err != nil {
		return fmt.Errorf("error serializing checkpoint (%d): %w", checkpointNum, err)
	}

	duration := time.Since(startTime)
	logger.Info().Float64("total_time_s", duration.Seconds()).Msgf("created checkpoint %d", checkpointNum)

	return nil
}

func cleanupCheckpoints(checkpointer *realWAL.Checkpointer, checkpointsToKeep int) error {
	// Don't list checkpoints if we keep them all
	if checkpointsToKeep == 0 {
		return nil
	}
	checkpoints, err := checkpointer.Checkpoints()
	if err != nil {
		return fmt.Errorf("cannot list checkpoints: %w", err)
	}
	if len(checkpoints) > int(checkpointsToKeep) {
		// if condition guarantees this never fails
		checkpointsToRemove := checkpoints[:len(checkpoints)-int(checkpointsToKeep)]

		for _, checkpoint := range checkpointsToRemove {
			err := checkpointer.RemoveCheckpoint(checkpoint)
			if err != nil {
				return fmt.Errorf("cannot remove checkpoint %d: %w", checkpoint, err)
			}
		}
	}
	return nil
}

// processTrieUpdate writes trie update to WAL, updates activeSegmentNum,
// and gets tries from trieQueue for checkpointing if needed.
// It also sends WAL update result and waits for trie update completion.
func (c *Compactor) processTrieUpdate(
	update *WALTrieUpdate,
	trieQueue *realWAL.TrieQueue,
	activeSegmentNum int,
	nextCheckpointNum int,
) (
	_activeSegmentNum int,
	checkpointNum int,
	checkpointTries []*trie.MTrie,
) {

	// RecordUpdate returns the segment number the record was written to.
	// Returned segment number (>= 0) can be
	// - the same as previous segment number (same segment), or
	// - incremented by 1 from previous segment number (new segment)
	segmentNum, skipped, updateErr := c.wal.RecordUpdate(update.Update)

	// Send result of WAL update
	update.ResultCh <- updateErr

	// This ensures that updated trie matches WAL update.
	defer func() {
		// Wait for updated trie
		trie := <-update.TrieCh
		if trie == nil {
			c.logger.Error().Msg("compactor failed to get updated trie")
			return
		}

		trieQueue.Push(trie)
	}()

	if activeSegmentNum == -1 {
		// Recover from failure to get active segment number at initialization.
		return segmentNum, -1, nil
	}

	if updateErr != nil || skipped || segmentNum == activeSegmentNum {
		return activeSegmentNum, -1, nil
	}

	// In the remaining code: segmentNum > activeSegmentNum

	// active segment is finalized.

	// Check new segment number is incremented by 1
	if segmentNum != activeSegmentNum+1 {
		c.logger.Error().Msg(fmt.Sprintf("compactor got unexpected new segment numer %d, want %d", segmentNum, activeSegmentNum+1))
	}

	// Update activeSegmentNum
	prevSegmentNum := activeSegmentNum
	activeSegmentNum = segmentNum

	if nextCheckpointNum > prevSegmentNum {
		// Not enough segments for checkpointing
		return activeSegmentNum, -1, nil
	}

	// In the remaining code: nextCheckpointNum == prevSegmentNum

	// Enough segments are created for checkpointing

	// Get tries from checkpoint queue.
	// At this point, checkpoint queue contains tries up to
	// last update (logged as last record in finalized segment)
	// It doesn't include trie for this update
	// until updated trie is received and added to trieQueue.
	tries := trieQueue.Tries()

	checkpointNum = nextCheckpointNum

	return activeSegmentNum, checkpointNum, tries
}
