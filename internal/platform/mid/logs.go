package mid

import (
	"context"
	"fmt"
	"time"

	"github.com/agkmw/reddit-clone/internal/platform/logger"
	"github.com/agkmw/reddit-clone/internal/platform/web"
)

func Logs(
	ctx context.Context,
	log *logger.Logger,
	hdl Handler,
	remoteAddr,
	method,
	path,
	query string,
) error {
	t := web.GetTracer(ctx)

	if query != "" {
		path = fmt.Sprintf("%s?%s", path, query)
	}

	log.Info(
		ctx,
		"request started",
		"remoteAddr", remoteAddr,
		"method", method,
		"path", path,
	)

	err := hdl(ctx)

	dur := time.Since(t.Now)

	fields := []any{
		"remoteAddr", remoteAddr,
		"method", method,
		"path", path,
		"status", t.StatusCode,
		"since", dur.String(),
	}

	if dur > 500*time.Millisecond {
		log.Warn(ctx, "request completed (slow)", fields...)
	} else {
		log.Info(ctx, "request completed", fields...)
	}

	return err
}
