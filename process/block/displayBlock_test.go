package block

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func createGenesisBlock(shardId uint32) *block.Header {
	rootHash := []byte("roothash")
	return &block.Header{
		Nonce:           0,
		Round:           0,
		Signature:       rootHash,
		RandSeed:        rootHash,
		PrevRandSeed:    rootHash,
		ShardID:         shardId,
		PubKeysBitmap:   rootHash,
		RootHash:        rootHash,
		PrevHash:        rootHash,
		MetaBlockHashes: [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
}

func createMockArgsTransactionCounter() ArgsTransactionCounter {
	return ArgsTransactionCounter{
		AppStatusHandler: &statusHandler.AppStatusHandlerStub{},
		Hasher:           &testscommon.HasherStub{},
		Marshalizer:      &testscommon.MarshalizerMock{},
		ShardID:          0,
	}
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenHasherIsNil(t *testing.T) {
	t.Parallel()

	args := createMockArgsTransactionCounter()
	args.Hasher = nil
	txCounter, err := NewTransactionCounter(args)

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenMarshalizerIsNil(t *testing.T) {
	t.Parallel()

	args := createMockArgsTransactionCounter()
	args.Marshalizer = nil
	txCounter, err := NewTransactionCounter(args)

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestDisplayBlock_NewTransactionCounterShouldErrWhenAppStatusHandlerIsNil(t *testing.T) {
	t.Parallel()

	args := createMockArgsTransactionCounter()
	args.AppStatusHandler = nil
	txCounter, err := NewTransactionCounter(args)

	assert.Nil(t, txCounter)
	assert.Equal(t, process.ErrNilAppStatusHandler, err)
}

func TestDisplayBlock_NewTransactionCounterShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgsTransactionCounter()
	txCounter, err := NewTransactionCounter(args)

	assert.NotNil(t, txCounter)
	assert.Nil(t, err)
}

func TestDisplayBlock_DisplayMetaHashesIncluded(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	header := createGenesisBlock(0)
	args := createMockArgsTransactionCounter()
	txCounter, _ := NewTransactionCounter(args)
	lines := txCounter.displayMetaHashesIncluded(
		shardLines,
		header,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(header.MetaBlockHashes), len(lines))
}

func TestDisplayBlock_DisplayTxBlockBody(t *testing.T) {
	t.Parallel()

	shardLines := make([]*display.LineData, 0)
	body := &block.Body{}
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")},
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)
	args := createMockArgsTransactionCounter()
	txCounter, _ := NewTransactionCounter(args)
	lines := txCounter.displayTxBlockBody(
		shardLines,
		&block.Header{},
		body,
	)

	assert.NotNil(t, lines)
	assert.Equal(t, len(miniblock.TxHashes), len(lines))
}

func TestDisplayBlock_GetConstructionStateAsString(t *testing.T) {
	miniBlockHeader := &block.MiniBlockHeader{}

	_ = miniBlockHeader.SetConstructionState(int32(block.Proposed))
	str := getConstructionStateAsString(miniBlockHeader)
	assert.Equal(t, "Proposed_", str)

	_ = miniBlockHeader.SetConstructionState(int32(block.PartialExecuted))
	str = getConstructionStateAsString(miniBlockHeader)
	assert.Equal(t, "Partial_", str)

	_ = miniBlockHeader.SetConstructionState(int32(block.Final))
	str = getConstructionStateAsString(miniBlockHeader)
	assert.Equal(t, "", str)
}

func TestDisplayBlock_ConcurrencyTestForTotalTxs(t *testing.T) {
	t.Parallel()

	numCalls := 100
	wg := sync.WaitGroup{}
	wg.Add(numCalls)

	args := createMockArgsTransactionCounter()
	txCounter, _ := NewTransactionCounter(args)

	mbh1 := block.MiniBlockHeader{}
	_ = mbh1.SetIndexOfLastTxProcessed(0)
	_ = mbh1.SetIndexOfLastTxProcessed(37)
	header := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{mbh1},
	}

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)
			defer wg.Done()

			switch idx % 2 {
			case 0:
				txCounter.headerReverted(header)
			case 1:
				txCounter.headerExecuted(header)
			}
		}(i)
	}

	wg.Wait()
}

func TestTransactionCounter_HeaderExecutedAndReverted(t *testing.T) {
	t.Parallel()

	args := createMockArgsTransactionCounter()

	mbhPeer := block.MiniBlockHeader{}
	_ = mbhPeer.SetTypeInt32(int32(block.PeerBlock))
	_ = mbhPeer.SetIndexOfFirstTxProcessed(0)
	_ = mbhPeer.SetIndexOfLastTxProcessed(99)

	mbhRwd := block.MiniBlockHeader{}
	_ = mbhRwd.SetTypeInt32(int32(block.RewardsBlock))
	_ = mbhRwd.SetIndexOfFirstTxProcessed(0)
	_ = mbhRwd.SetIndexOfLastTxProcessed(199)

	mbhScheduledFromShard0 := block.MiniBlockHeader{}
	_ = mbhScheduledFromShard0.SetTypeInt32(int32(block.TxBlock))
	_ = mbhScheduledFromShard0.SetProcessingType(int32(block.Scheduled))
	_ = mbhScheduledFromShard0.SetIndexOfFirstTxProcessed(0)
	_ = mbhScheduledFromShard0.SetIndexOfLastTxProcessed(399)

	mbhScheduledFromShard1 := block.MiniBlockHeader{
		SenderShardID: 1,
	}
	_ = mbhScheduledFromShard1.SetTypeInt32(int32(block.TxBlock))
	_ = mbhScheduledFromShard1.SetProcessingType(int32(block.Scheduled))
	_ = mbhScheduledFromShard1.SetIndexOfFirstTxProcessed(0)
	_ = mbhScheduledFromShard1.SetIndexOfLastTxProcessed(499)

	t.Run("headerExecuted", func(t *testing.T) {
		t.Parallel()

		txCounter, _ := NewTransactionCounter(args)
		t.Run("nil header should not panic", func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					assert.Fail(t, fmt.Sprintf("should not have panicked: %v", r))
				}
			}()

			txCounter.headerExecuted(nil)
		})
		t.Run("empty header", func(t *testing.T) {
			txCounter.totalTxs = 1000 // initial value
			txCounter.headerExecuted(&block.Header{})
			assert.Equal(t, uint64(1000), txCounter.totalTxs)
		})
		t.Run("header with peer miniblocks & rewards miniblocks", func(t *testing.T) {
			txCounter.totalTxs = 1000 // initial value

			blk := &block.Header{
				MiniBlockHeaders: []block.MiniBlockHeader{mbhPeer, mbhRwd},
			}

			txCounter.headerExecuted(blk)
			assert.Equal(t, uint64(1200), txCounter.totalTxs)
		})
		t.Run("header with scheduled from self and shard 1", func(t *testing.T) {
			txCounter.totalTxs = 1000 // initial value

			blk := &block.Header{
				MiniBlockHeaders: []block.MiniBlockHeader{mbhScheduledFromShard0, mbhScheduledFromShard1},
			}

			txCounter.headerExecuted(blk)
			assert.Equal(t, uint64(1500), txCounter.totalTxs)
		})
	})
	t.Run("headerReverted", func(t *testing.T) {
		t.Parallel()

		txCounter, _ := NewTransactionCounter(args)
		t.Run("nil header should not panic", func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil {
					assert.Fail(t, fmt.Sprintf("should not have panicked: %v", r))
				}
			}()

			txCounter.headerReverted(nil)
		})
		t.Run("empty header", func(t *testing.T) {
			txCounter.totalTxs = 1000 // initial value
			txCounter.headerReverted(&block.Header{})
			assert.Equal(t, uint64(1000), txCounter.totalTxs)
		})
		t.Run("header with peer miniblocks & rewards miniblocks", func(t *testing.T) {
			txCounter.totalTxs = 1000 // initial value
			blk := &block.Header{
				MiniBlockHeaders: []block.MiniBlockHeader{mbhPeer, mbhRwd},
			}

			txCounter.headerReverted(blk)
			assert.Equal(t, uint64(800), txCounter.totalTxs) // 1000 - 200
		})
		t.Run("header with scheduled from self and shard 1", func(t *testing.T) {
			txCounter.totalTxs = 1000 // initial value
			blk := &block.Header{
				MiniBlockHeaders: []block.MiniBlockHeader{mbhScheduledFromShard0, mbhScheduledFromShard1},
			}

			txCounter.headerReverted(blk)
			assert.Equal(t, uint64(500), txCounter.totalTxs) // 1000 - 500
		})
	})
	t.Run("headerExecuted then headerReverted", func(t *testing.T) {
		t.Parallel()

		txCounter, _ := NewTransactionCounter(args)
		txCounter.totalTxs = 1000 // initial value
		blk := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{mbhPeer, mbhRwd, mbhScheduledFromShard0, mbhScheduledFromShard1},
		}

		txCounter.headerExecuted(blk)
		assert.Equal(t, uint64(1700), txCounter.totalTxs)

		txCounter.headerReverted(blk)
		assert.Equal(t, uint64(1000), txCounter.totalTxs)
	})
}
