package interceptors_test

import (
	"bytes"
	"errors"
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewPeerBlockBodyInterceptor

func TestNewPeerBlockBodyInterceptor_WithNilParameterShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, err := interceptors.NewPeerBlockBodyInterceptor(
		nil,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, pbbi)
}

func TestNewPeerBlockBodyInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, err := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Nil(t, err)
	assert.NotNil(t, pbbi)
}

//------- ProcessReceivedMessage

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	assert.Equal(t, process.ErrNilMessage, pbbi.ProcessReceivedMessage(nil))
}

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageNilMessageDataShouldErr(t *testing.T) {
	t.Parallel()

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerMock{},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{}

	assert.Equal(t, process.ErrNilDataToProcess, pbbi.ProcessReceivedMessage(msg))
}

func TestPeerBlockBodyInterceptor_ValidateMarshalizerErrorsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	errMarshalizer := errors.New("marshalizer error")

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errMarshalizer
			},
		},
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	assert.Equal(t, errMarshalizer, pbbi.ProcessReceivedMessage(msg))
}

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageIntegrityFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	peerChangeBlock := block.NewInterceptedPeerBlockBody()
	peerChangeBlock.ShardID = uint32(0)
	peerChangeBlock.RootHash = []byte("root hash")
	peerChangeBlock.Changes = nil

	buff, _ := marshalizer.Marshal(peerChangeBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	assert.Equal(t, process.ErrNilPeerChanges, pbbi.ProcessReceivedMessage(msg))
}

func TestPeerBlockBodyInterceptor_ProcessReceivedMessageBlockShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	cache := &mock.CacherStub{}
	storer := &mock.StorerStub{
		HasCalled: func(key []byte) (b bool, e error) {
			return false, nil
		},
	}

	pbbi, _ := interceptors.NewPeerBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		mock.HasherMock{},
		mock.NewOneShardCoordinatorMock())

	peerChangeBlock := block.NewInterceptedPeerBlockBody()
	peerChangeBlock.ShardID = uint32(0)
	peerChangeBlock.RootHash = []byte("root hash")
	peerChangeBlock.Changes = []block2.PeerChange{
		{PubKey: []byte("pub key"), ShardIdDest: uint32(0)},
	}

	buff, _ := marshalizer.Marshal(peerChangeBlock)

	msg := &mock.P2PMessageMock{
		DataField: buff,
	}

	putInCacheWasCalled := false
	cache.PutCalled = func(key []byte, value interface{}) (evicted bool) {
		if bytes.Equal(key, mock.HasherMock{}.Compute(string(buff))) {
			putInCacheWasCalled = true
		}
		return false
	}

	assert.Nil(t, pbbi.ProcessReceivedMessage(msg))
	assert.True(t, putInCacheWasCalled)
}