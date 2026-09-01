package collectors

import (
	"errors"

	"github.com/aifia105/kubeguard/pkg/logger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
)

func HandleScanError(scan string, err error) error {
	if err == nil {
		return nil
	}

	var noMatchErr *meta.NoKindMatchError
	switch {
	case errors.As(err, &noMatchErr):
		logger.LogWarning("%s: API not registered in this cluster (likely not installed), skipping", scan)
	case apierrors.IsNotFound(err):
		logger.LogWarning("%s: resource not found, skipping", scan)
	case apierrors.IsServiceUnavailable(err):
		logger.LogWarning("%s: service unavailable, skipping", scan)
	default:
		return err
	}

	return nil
}
