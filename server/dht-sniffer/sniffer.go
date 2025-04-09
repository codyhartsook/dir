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
	"log"
	"time"
	_ "time"

	"golang.org/x/sync/errgroup"
	_ "golang.org/x/sync/errgroup"

	// "go.opentelemetry.io/otel/codes"
	"github.com/agntcy/dir/server/dht-sniffer/eventsource"
	e "github.com/agntcy/dir/server/dht-sniffer/eventsource"
	"github.com/agntcy/dir/server/dht-sniffer/handler"
	"github.com/ipfs-search/ipfs-search/instr"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/redis/go-redis/v9"
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

	*instr.Instrumentation
}

// New creates a new Sniffer based on a datastore, or returns an error.
func New(cfg *Config, ds datastore.Batching, i *instr.Instrumentation) (*Sniffer, error) {
	bus := eventbus.NewBus()

	es, err := eventsource.New(bus, ds)
	if err != nil {
		return nil, fmt.Errorf("failed to get eventsource: %w", err)
	}

	s := Sniffer{
		cfg:             cfg,
		es:              es,
		Instrumentation: i,
	}

	return &s, nil
}

// Batching returns the datastore wrapped with sniffing hooks.
func (s *Sniffer) Batching() datastore.Batching {
	return s.es.Batching()
}

func (s *Sniffer) consume(ctx context.Context, c chan e.EvtProviderPut) error {
	redisAddr := fmt.Sprintf("%s:%d", "localhost", 6379)
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Failed to connect to Redis at %s: %v\n", redisAddr, err)
		return err
	}

	go func() {
		// read from the channel and publish to the queue
		for {
			select {
			case <-ctx.Done():
				return
			case p := <-c:
				// Publish to Redis
				err = client.Publish(ctx, "messages", p.Agent).Err()
				if err != nil {
					fmt.Printf("Error publishing message: %v\n", err)
				} else {
					fmt.Printf("Published message: %s\n", string(p.Agent))
				}

				fmt.Println("agentBytes", string(p.Agent))
				// get the model given the agent
			}
		}
	}()

	return nil
}

func (s *Sniffer) subscribe(ctx context.Context, c chan e.EvtProviderPut) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.subscribe")
	// defer span.End()

	h := handler.New(c) // Ensure handler matches the eventsource package

	err := s.es.Subscribe(ctx, h.HandleFunc)
	// span.RecordError(err)
	// span.SetStatus(codes.Internal, err.Error())
	return err
}

func (s *Sniffer) iterate(ctx context.Context, sniffed chan e.EvtProviderPut) error {
	// Create error group and context
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error { return s.subscribe(ctx, sniffed) })
	errg.Go(func() error { return s.consume(ctx, sniffed) })

	// Wait until all contexts are closed, then return *first* error
	err := errg.Wait()

	return err
}

// Sniff starts sniffing until the context is closed - it restarts itself on intermittant errors.
func (s *Sniffer) Sniff(ctx context.Context) error {
	// ctx, span := s.Tracer.Start(ctx, "sniffer.Sniff")
	// defer span.End()

	sniffed := make(chan e.EvtProviderPut, s.cfg.BufferSize)

	for {
		err := s.iterate(ctx, sniffed)

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
