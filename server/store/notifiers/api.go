package notifiers

import (
	"context"
	"io"

	coretypes "github.com/agntcy/dir/api/core/v1alpha1"
)

type Notifier interface {
	NotifyPush(ctx context.Context, ref *coretypes.ObjectRef, reader io.Reader) error
	NotifyDelete(ctx context.Context, ref *coretypes.ObjectRef) error
}
