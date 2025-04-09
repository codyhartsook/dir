package factory

import (
	"context"
	"fmt"

	sniffer "github.com/agntcy/dir/server/dht-sniffer"

	"github.com/ipfs/go-datastore"
)

// Start initialises a sniffer and all its dependencies and launches it in a goroutine, returning a wrapped context
// and datastore, which should replace the original ones, or an error from initialisation.
func Start(ctx context.Context, ds datastore.Batching) (context.Context, datastore.Batching, error) {
	cfg := sniffer.DefaultConfig()

	// Create context which can be canceled by sniffer so as to propagate failure from sniffer goroutine.
	ctx, cancel := context.WithCancel(ctx)

	s, err := sniffer.New(cfg, ds, nil)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	// Use batched datastore
	ds = s.Batching()

	// Start sniffer
	go func() {
		// Cancel parent context when done
		defer cancel()

		err = s.Sniff(ctx)
		fmt.Printf("Sniffer exited: %s\n", err)
	}()

	return ctx, ds, nil
}
