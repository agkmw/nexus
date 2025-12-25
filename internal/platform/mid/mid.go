package mid

import (
	"context"
)

type Handler func(ctx context.Context) error
