package handler

import (
	"context"

	"github.com/agntcy/dir/server/dht-sniffer/eventsource"
	e "github.com/agntcy/dir/server/dht-sniffer/eventsource"
	"github.com/ipfs-search/ipfs-search/instr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Handler handles EvtProviderPut events by writing Provider's to a channel.
type Handler struct {
	providers chan<- e.EvtProviderPut
	*instr.Instrumentation
}

// New returns a new handler, writing Provider's to providers.
func New(providers chan<- e.EvtProviderPut) Handler {
	return Handler{
		providers:       providers,
		Instrumentation: instr.New(),
	}
}

// HandleFunc writes a Provider to the Handler's providers channel for every EvtProviderPut it is called with.
func (h *Handler) HandleFunc(ctx context.Context, e eventsource.EvtProviderPut) error {
	//fmt.Println("handler func called")
	ctx = trace.ContextWithRemoteSpanContext(ctx, e.SpanContext)
	ctx, span := h.Tracer.Start(ctx, "handler.HandleFunc", trace.WithAttributes(
		attribute.Stringer("peerid", e.PeerID),
	), trace.WithSpanKind(trace.SpanKindConsumer))

	defer span.End()

	// send the provider to the channel
	h.providers <- e

	return nil
}
