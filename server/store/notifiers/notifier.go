// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

//nolint:wrapcheck
package notifiers

import (
	"bytes"
	"context"
	"fmt"
	"io"

	coretypes "github.com/agntcy/dir/api/core/v1alpha1"
	"github.com/agntcy/dir/server/store/notifiers/webhook"
	"github.com/agntcy/dir/server/types"
)

type Publisher string

const (
	WEBHOOK = Publisher("webhook")
	SQL     = Publisher("sql")
	REDIS   = Publisher("redis")
	NATS    = Publisher("nats")
)

type store struct {
	source   types.StoreAPI
	notifier Notifier
}

// TODO: benchmark notifiers vs direct store access
func Wrap(opts types.APIOptions, source types.StoreAPI) types.StoreAPI {
	var notifier Notifier
	fmt.Println("Using notifier:", opts.Config().Notifier)
	switch Publisher := Publisher(opts.Config().Notifier); Publisher {
	case WEBHOOK:
		notifier = webhook.New(opts)
	case SQL:
		fmt.Println("Using SQL publisher")
	case REDIS:
		fmt.Println("Using Redis publisher")
	case NATS:
		fmt.Println("Using NATS publisher")
	default:
		fmt.Println("Using default publisher")
	}

	return &store{
		source:   source,
		notifier: notifier,
	}
}

func (s *store) Push(ctx context.Context, ref *coretypes.ObjectRef, reader io.Reader) (*coretypes.ObjectRef, error) {
	// Read all data into a buffer
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// push data
	ref, err = s.source.Push(ctx, ref, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	if s.notifier != nil {
		// Notify the notifier in a goroutine
		go func() {
			s.notifier.NotifyPush(ctx, ref, bytes.NewReader(data))
		}()
	}

	return ref, nil
}

func (s *store) Pull(ctx context.Context, ref *coretypes.ObjectRef) (io.ReadCloser, error) {
	return s.source.Pull(ctx, ref)
}

func (s *store) Lookup(ctx context.Context, ref *coretypes.ObjectRef) (*coretypes.ObjectRef, error) {
	// fetch from source
	sourceRef, err := s.source.Lookup(ctx, ref)
	if err != nil {
		return nil, err
	}

	return sourceRef, nil
}

func (s *store) Delete(ctx context.Context, ref *coretypes.ObjectRef) error {
	// delete
	if err := s.source.Delete(ctx, ref); err != nil {
		return err
	}

	if s.notifier != nil {
		s.notifier.NotifyDelete(ctx, ref)
	}

	return nil
}
