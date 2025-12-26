package web

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func RateLimit(
	enabled bool,
	rps float64,
	burst int,
) Middleware {
	mid := func(handler Handler) Handler {
		type client struct {
			limiter  *rate.Limiter
			lastSeen time.Time
		}

		var (
			mu      sync.Mutex
			clients = make(map[string]*client)
		)

		go func() {
			time.Sleep(time.Minute)

			mu.Lock()
			{
				for ip, client := range clients {
					if time.Since(client.lastSeen) >= 3*time.Minute {
						delete(clients, ip)
					}
				}
			}
			mu.Unlock()
		}()

		hdl := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if enabled {
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					return ServerErrorResponse(ctx, w)
				}

				mu.Lock()
				{
					if _, found := clients[ip]; !found {
						clients[ip] = &client{
							limiter: rate.NewLimiter(rate.Limit(rps), burst),
						}
					}

					clients[ip].lastSeen = time.Now()

					if !clients[ip].limiter.Allow() {
						mu.Unlock()
						return RateLimitExceededResponse(ctx, w)
					}
				}
				mu.Unlock()
			}

			return handler(ctx, w, r)
		}

		return hdl
	}

	return mid
}
