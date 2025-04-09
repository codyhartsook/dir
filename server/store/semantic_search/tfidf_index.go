package searchindex

import (
	"context"
	"encoding/json"
	"io"
	"log"

	coretypes "github.com/agntcy/dir/api/core/v1alpha1"
	routetypes "github.com/agntcy/dir/api/routing/v1alpha1"
	"github.com/agntcy/dir/server/store/semantic_search/config"
	"github.com/agntcy/dir/server/types"
	"github.com/blevesearch/bleve"
)

type BleveIndex struct {
	index bleve.Index
}

func New(cfg config.Config) (types.SearchIndexAPI, error) {
	mapping := bleve.NewIndexMapping()

	// check if the index already exists at dir
	index, err := bleve.Open(cfg.Dir)
	if err == nil {
		return &BleveIndex{index: index}, nil
	}
	index, err = bleve.New(cfg.Dir, mapping)
	if err != nil {
		return nil, err
	}

	return &BleveIndex{index: index}, nil
}

func (b *BleveIndex) Index(ctx context.Context, ref *coretypes.ObjectRef, record io.Reader) error {
	// Read the data from the record into a coretypes.Object
	var data coretypes.Object
	if err := json.NewDecoder(record).Decode(&data); err != nil {
		log.Printf("Error decoding JSON: %v", err)
	}
	if data.GetAgent() == nil {
		log.Printf("Error: Agent is not set in the object")
		return nil
	}
	if err := b.index.Index(ref.Digest, data.GetAgent()); err != nil {
		return err
	}
	return nil
}

func (b *BleveIndex) Delete(ctx context.Context, ref *coretypes.ObjectRef) error {
	if err := b.index.Delete(ref.Digest); err != nil {
		return err
	}
	return nil
}

func (b *BleveIndex) Search(ctx context.Context, req *routetypes.SearchRequest) (io.ReadCloser, error) {
	searchLabels := req.GetLabels()
	k := req.GetSize()

	// TODO: join the search labels or update the proto to accept a single query
	query := searchLabels[0]

	q := bleve.NewMatchQuery(query)
	searchRequest := bleve.NewSearchRequestOptions(q, int(k), 0, false)
	search := bleve.NewSearchRequest(searchRequest.Query)

	// see if only the first k results are needed
	if k > 0 {
		search.Size = int(k)
	}
	result, err := b.index.Search(search)
	if err != nil {
		log.Printf("Error searching index: %v", err)
		return nil, err
	}

	for _, hit := range result.Hits {
		log.Printf("Found hit: %s", hit.ID)
	}

	return nil, nil
}
