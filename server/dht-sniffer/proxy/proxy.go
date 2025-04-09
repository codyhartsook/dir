package proxy

import (
	"github.com/agntcy/dir/server/dht-sniffer/proxy/batch"
	"github.com/ipfs/go-datastore"
)

// New wraps a datastore in a proxy calling afterPut after every Put() operation.
func New(ds datastore.Batching, afterPut AfterPutFunc) datastore.Batching {
	afterBatch := func(b datastore.Batch, err error) (datastore.Batch, error) {
		return batch.NewBatch(b, batch.WithAfterPut(batch.AfterPutFunc(afterPut))), err
	}

	return NewBatching(ds, WithAfterPut(afterPut), WithAfterBatch(afterBatch))
}
