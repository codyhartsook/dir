// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"io"

	coretypes "github.com/agntcy/dir/api/core/v1alpha1"
	routetypes "github.com/agntcy/dir/api/routing/v1alpha1"
)

type SearchIndexAPI interface {
	// Index adds an object to the search index.
	Index(context.Context, *coretypes.ObjectRef, io.Reader) error
	// delete removes an object from the search index.
	Delete(context.Context, *coretypes.ObjectRef) error
	// Search returns the top k objects that match the query.
	Search(context.Context, *routetypes.SearchRequest) (io.ReadCloser, error)
}
