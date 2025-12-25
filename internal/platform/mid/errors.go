package mid

import (
	"context"

	"github.com/agkmw/reddit-clone/internal/platform/errs"
	"github.com/agkmw/reddit-clone/internal/platform/logger"
)

func Errors(
	ctx context.Context,
	log *logger.Logger,
	hdl Handler,
) error {
	err := hdl(ctx)
	if nil == err {
		return nil
	}

	log.Error(ctx, "request failed", "error", "error", err)

	if e, ok := errs.Get(err); ok {
		return e
	}

	return errs.New(errs.Unknown, err, nil)
}
