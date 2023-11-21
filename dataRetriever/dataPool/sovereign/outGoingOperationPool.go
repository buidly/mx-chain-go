package sovereign

import (
	"sort"
	"sync"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("outgoing-operations-pool")

type OutGoingOperationsData struct {
	Hash []byte
	Data map[string][]byte
}

type cacheEntry struct {
	data     *OutGoingOperationsData
	expireAt time.Time
}

// This is a cache which stores outgoing txs data at their specified hash.
// Each entry in cache has an expiry time. We should delete entries from this cache once the confirmation from the notifier
// is received that the outgoing operation has been sent to main chain.
// An unconfirmed operation is a tx data operation which has been stored in cache for longer than the time to wait for
// unconfirmed outgoing operations.
// The leader of the next round should check if there are any unconfirmed operations and try to resend them.

// TODO: [TBD] here a mechanism to check for unconfirmed operations inside the network as well in order not to double spend?
// E.g.: are there truly unconfirmed operations or the leader is having trouble with his notifier and didn't receive the confirmation?
type outGoingOperationsPool struct {
	mutex   sync.RWMutex
	timeout time.Duration
	cache   map[string]cacheEntry
}

// NewOutGoingOperationPool creates a new outgoing operation pool able to store data with an expiry time
func NewOutGoingOperationPool(expiryTime time.Duration) *outGoingOperationsPool {
	log.Debug("NewOutGoingOperationPool", "time to wait for unconfirmed outgoing operations", expiryTime)

	return &outGoingOperationsPool{
		timeout: expiryTime,
		cache:   map[string]cacheEntry{},
	}
}

// Add adds the outgoing txs data at the specified hash in the internal cache
func (op *outGoingOperationsPool) Add(data *OutGoingOperationsData) {
	hashStr := string(data.Hash)

	op.mutex.Lock()
	defer op.mutex.Unlock()

	if _, exists := op.cache[hashStr]; exists {
		return
	}

	op.cache[hashStr] = cacheEntry{
		data:     data,
		expireAt: time.Now().Add(op.timeout),
	}
}

// Get returns the outgoing txs data at the specified hash
func (op *outGoingOperationsPool) Get(hash []byte) *OutGoingOperationsData {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	return op.cache[string(hash)].data
}

// Delete removes the outgoing tx data at the specified hash
func (op *outGoingOperationsPool) Delete(hash []byte) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	delete(op.cache, string(hash))
}

// GetUnconfirmedOperations returns a list of unconfirmed operations.
// An unconfirmed operation is a tx data operation which has been stored in cache for longer
// than the time to wait for unconfirmed outgoing operations.
// Returned list is sorted based on expiry time.
func (op *outGoingOperationsPool) GetUnconfirmedOperations() []*OutGoingOperationsData {
	expiredEntries := make([]cacheEntry, 0)

	op.mutex.Lock()
	for _, entry := range op.cache {
		if time.Now().After(entry.expireAt) {
			expiredEntries = append(expiredEntries, entry)
		}
	}
	op.mutex.Unlock()

	sort.Slice(expiredEntries, func(i, j int) bool {
		return expiredEntries[i].expireAt.Before(expiredEntries[j].expireAt)
	})

	ret := make([]*OutGoingOperationsData, len(expiredEntries))
	for i, entry := range expiredEntries {
		ret[i] = entry.data
	}

	return ret
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outGoingOperationsPool) IsInterfaceNil() bool {
	return op == nil
}
