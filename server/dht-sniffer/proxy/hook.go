package proxy

import (
	"context"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

// Datastore is a wrapper for a datastore that adds optional before and after hooks into it's methods.
type Datastore struct {
	ds      datastore.Datastore
	options Options
}

// NewDatastore wraps a datastore.Datastore datastore and adds optional before and after hooks into it's methods.
func NewDatastore(ds datastore.Datastore, options ...Option) *Datastore {
	opts := Options{}
	opts.Apply(options...)
	return &Datastore{ds: ds, options: opts}
}

// Put stores the object `value` named by `key`, it calls OnBeforePut and OnAfterPut hooks.
func (hds *Datastore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	if hds.options.BeforePut != nil {
		key, value = hds.options.BeforePut(key, value)
	}
	err := hds.ds.Put(ctx, key, value)
	if hds.options.AfterPut != nil {
		err = hds.options.AfterPut(key, value, err)
	}
	return err
}

// Delete removes the value for given `key`, it calls OnBeforeDelete and OnAfterDelete hooks.
func (hds *Datastore) Delete(ctx context.Context, key datastore.Key) error {
	if hds.options.BeforeDelete != nil {
		key = hds.options.BeforeDelete(key)
	}
	err := hds.ds.Delete(ctx, key)
	if hds.options.AfterDelete != nil {
		err = hds.options.AfterDelete(key, err)
	}
	return err
}

// Get retrieves the object `value` named by `key`, it calls OnBeforeGet and OnAfterGet hooks.
func (hds *Datastore) Get(ctx context.Context, key datastore.Key) ([]byte, error) {
	if hds.options.BeforeGet != nil {
		key = hds.options.BeforeGet(key)
	}
	value, err := hds.ds.Get(ctx, key)
	if hds.options.AfterGet != nil {
		value, err = hds.options.AfterGet(key, value, err)
	}
	return value, err
}

// Has returns whether the `key` is mapped to a `value`.
func (hds *Datastore) Has(ctx context.Context, key datastore.Key) (bool, error) {
	if hds.options.BeforeHas != nil {
		key = hds.options.BeforeHas(key)
	}
	exists, err := hds.ds.Has(ctx, key)
	if hds.options.AfterHas != nil {
		exists, err = hds.options.AfterHas(key, exists, err)
	}
	return exists, err
}

// GetSize returns the size of the `value` named by `key`.
func (hds *Datastore) GetSize(ctx context.Context, key datastore.Key) (int, error) {
	return hds.ds.GetSize(ctx, key)
}

// Query searches the datastore and returns a query result, it calls OnBeforeQuery and OnAfterQuery hooks.
func (hds *Datastore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	if hds.options.BeforeQuery != nil {
		q = hds.options.BeforeQuery(q)
	}
	res, err := hds.ds.Query(ctx, q)
	if hds.options.AfterQuery != nil {
		res, err = hds.options.AfterQuery(q, res, err)
	}
	return res, err
}

// Sync guarantees that any Put or Delete calls under prefix that returned
// before Sync(prefix) was called will be observed after Sync(prefix)
// returns, even if the program crashes.
func (hds *Datastore) Sync(ctx context.Context, prefix datastore.Key) error {
	return hds.ds.Sync(ctx, prefix)
}

// Close closes the underlying datastore
func (hds *Datastore) Close() error {
	return hds.ds.Close()
}
