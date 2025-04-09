package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/huggingface"
)

// Huggingface is the embedder using the Huggingface hub api.
type Huggingface struct {
	client *huggingface.LLM
	Model  string
	Task   string

	StripNewLines bool
	BatchSize     int
}

var _ embeddings.Embedder = &Huggingface{}

func NewHuggingface(opts ...Option) (*Huggingface, error) {
	v, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (e *Huggingface) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	batchedTexts := embeddings.BatchTexts(
		embeddings.MaybeRemoveNewLines(texts, e.StripNewLines),
		e.BatchSize,
	)

	emb := make([][]float32, 0, len(texts))
	for _, batch := range batchedTexts {
		curBatchEmbeddings, err := e.client.CreateEmbedding(ctx, batch, e.Model, e.Task)
		if err != nil {
			return nil, err
		}
		emb = append(emb, curBatchEmbeddings...)
	}

	return emb, nil
}

func (e *Huggingface) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if e.StripNewLines {
		text = strings.ReplaceAll(text, "\n", " ")
	}

	emb, err := e.client.CreateEmbedding(ctx, []string{text}, e.Model, e.Task)
	if err != nil {
		return nil, err
	}

	return emb[0], nil
}

func TestHuggingfaceEmbeddings(t *testing.T) {
	t.Parallel()

	if huggingfaceKey := os.Getenv("HUGGINGFACEHUB_API_TOKEN"); huggingfaceKey == "" {
		t.Skip("HUGGINGFACEHUB_API_TOKEN not set")
	}
	e, err := NewHuggingface()
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world!")
	require.NoError(t, err)

	embeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world", "The world is ending", "good bye"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
}
