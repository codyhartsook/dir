/*
Package sniffer contains sniffer components which can be wired into a libp2p dht node by proxying the datastore.

The canonical implementation thereof can be found in: https://github.com/ipfs-search/ipfs-sniffer

If all you want is to sniff CID's, you probably want to use a `factory` to facilitate the setup of the sniffer.
*/
package sniffer

import (
	"context"
	_ "context"
	"fmt"
	_ "fmt"
	"log"
	_ "log"
	"time"
	_ "time"

	"golang.org/x/sync/errgroup"
	_ "golang.org/x/sync/errgroup"

	// "go.opentelemetry.io/otel/codes"
	"github.com/agntcy/dir/server/dht-sniffer/eventsource"
	"github.com/agntcy/dir/server/dht-sniffer/handler"
	"github.com/ipfs-search/ipfs-search/components/queue"
	"github.com/ipfs-search/ipfs-search/instr"
	t "github.com/ipfs-search/ipfs-search/types"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
)

type Config struct {
	LastSeenExpiration time.Duration // Expiration time for the last-seen resources
	LastSeenPruneLen   int           // Cleanup expired resources from the last-seen
	LoggerTimeout      time.Duration // Throw timeout error when no log messages arrive
	BufferSize         uint          // Size of the channels buffering between yielder, filter and adder
}

// DefaultConfig returns the default configuration for a Sniffer.
func DefaultConfig() *Config {
	return &Config{
		LastSeenExpiration: 60 * time.Duration(time.Minute),
		LastSeenPruneLen:   32768,
		LoggerTimeout:      60 * time.Duration(time.Second),
		BufferSize:         512,
	}
}

// Sniffer allows sniffing Batching datastore's events, effectively allowing sniffing of the IPFS DHT.
// To effectively use the Sniffer, the proxied datastore needs to be acquired by calling `Batching()` on the Sniffer.
type Sniffer struct {
	cfg *Config
	es  eventsource.EventSource
	pub queue.PublisherFactory

	*instr.Instrumentation
}

// New creates a new Sniffer based on a datastore, or returns an error.
func New(cfg *Config, ds datastore.Batching, pub queue.PublisherFactory, i *instr.Instrumentation) (*Sniffer, error) {
	bus := eventbus.NewBus()
	fmt.Println("bus", bus)

	es, err := eventsource.New(bus, ds)
	if err != nil {
		return nil, fmt.Errorf("failed to get eventsource: %w", err)
	}

	s := Sniffer{
		cfg:             cfg,
		es:              es,
		pub:             pub,
		Instrumentation: i,
	}

	return &s, nil
}

// Batching returns the datastore wrapped with sniffing hooks.
func (s *Sniffer) Batching() datastore.Batching {
	return s.es.Batching()
}

func (s *Sniffer) subscribe(ctx context.Context, c chan<- t.Provider) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.subscribe")
	// defer span.End()

	h := handler.New(c) // Ensure handler matches the eventsource package

	err := s.es.Subscribe(ctx, h.HandleFunc)
	// span.RecordError(err)
	// span.SetStatus(codes.Internal, err.Error())
	return err
}

func (s *Sniffer) filter(ctx context.Context, in <-chan t.Provider, out chan<- t.Provider) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.filter")
	// defer span.End()

	//lastSeenFilter := filters.NewLastSeenFilter(s.cfg.LastSeenExpiration, s.cfg.LastSeenPruneLen)
	//cidFilter := filters.NewCidFilter()
	//mutliFilter := filters.NewMultiFilter(lastSeenFilter, cidFilter)
	//f := filter.New(mutliFilter, in, out)

	//err := f.Filter(ctx)
	// span.RecordError(err)
	// span.SetStatus(codes.Internal, err.Error())
	return nil
}

func (s *Sniffer) queue(ctx context.Context, c <-chan t.Provider) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.Queue")
	// defer span.End()

	/*publisher, err := s.pub.NewPublisher(ctx)
	if err != nil {
		return err
	}

	q := queuer.New(publisher, c)

	err = q.Queue(ctx)*/

	// create a listener in the background
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case p, ok := <-c:
				if !ok {
					return
				}
				fmt.Println("Queueing provider:", p)
			}
		}
	}()

	// span.RecordError(err)
	// span.SetStatus(codes.Internal, err.Error())
	return nil
}

func (s *Sniffer) iterate(ctx context.Context, sniffed, filtered chan t.Provider) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.iterate")
	// defer span.End()

	// Create error group and context
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error { return s.subscribe(ctx, sniffed) })
	errg.Go(func() error { return s.filter(ctx, sniffed, filtered) })
	errg.Go(func() error { return s.queue(ctx, filtered) })

	// Wait until all contexts are closed, then return *first* error
	err := errg.Wait()

	// span.RecordError(err)
	// span.SetStatus(codes.Internal, err.Error())

	return err
}

// Sniff starts sniffing until the context is closed - it restarts itself on intermittant errors.
func (s *Sniffer) Sniff(ctx context.Context) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.Sniff")
	// defer span.End()

	sniffed := make(chan t.Provider, s.cfg.BufferSize)
	filtered := make(chan t.Provider, s.cfg.BufferSize)

	for {
		err := s.iterate(ctx, sniffed, filtered)

		// Closing the parent context should cause a return, other errors cause a restart
		if err := ctx.Err(); err != nil {
			log.Printf("Parent context closed with error '%s', returning error", err)
			// span.RecordError(err)
			// span.SetStatus(codes.Internal, err.Error())
			return err
		}

		log.Printf("Wait group exited with error '%s', restarting", err)

		// TODO: Add circuit breaker here
		log.Printf("Stubbornly restarting in 1s")
		time.Sleep(time.Second)
	}
}
