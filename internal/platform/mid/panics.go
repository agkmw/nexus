package mid

import (
	"context"
	"fmt"

	"github.com/agkmw/reddit-clone/internal/platform/errs"
)

func Panics(ctx context.Context, hdl Handler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e := fmt.Errorf("%v", r)
			err = errs.New(errs.Internal, e, nil)
		}
	}()

	return hdl(ctx)
}
