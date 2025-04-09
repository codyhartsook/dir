package webhook

import (
	"context"
	"io"
	"net/http"

	coretypes "github.com/agntcy/dir/api/core/v1alpha1"
	"github.com/agntcy/dir/server/types"
	"github.com/agntcy/dir/utils/logging"
)

var (
	logger = logging.Logger("notifers/webhook")
)

type Webhook struct {
	// URL to send the webhook to
	URL string `json:"url,omitempty" mapstructure:"url"`
}

func New(opts types.APIOptions) *Webhook {
	// Get the webhook URL from the config
	url := opts.Config().Webhook
	return &Webhook{
		URL: url,
	}
}

func (w *Webhook) NotifyPush(ctx context.Context, ref *coretypes.ObjectRef, reader io.Reader) error {
	req, err := http.NewRequest("POST", w.URL, reader)
	if err != nil {
		logger.Error("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("Error sending request:", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Warn("Webhook returned non-200 status code:", resp.StatusCode)
	}

	return nil
}

func (w *Webhook) NotifyDelete(ctx context.Context, ref *coretypes.ObjectRef) error {
	req, err := http.NewRequest("DELETE", w.URL, nil)
	if err != nil {
		logger.Error("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("Error sending request:", err)
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Warn("Webhook returned non-200 status code:", resp.StatusCode)
	}
	return nil
}
