package web

type Middleware func(handler Handler) Handler

func wrapMiddleware(middlewares []Middleware, handler Handler) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mid := middlewares[i]
		if mid != nil {
			handler = mid(handler)
		}
	}

	return handler
}
