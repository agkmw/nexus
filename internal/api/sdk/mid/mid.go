package mid

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/agkmw/reddit-clone/internal/app/sdk/errs"
	"github.com/agkmw/reddit-clone/internal/platform/logger"
	"github.com/agkmw/reddit-clone/internal/platform/web"
	"golang.org/x/time/rate"
)

var httpStatus [8]int

func init() {
	httpStatus[errs.Internal.Value()] = http.StatusInternalServerError
	httpStatus[errs.BadRequest.Value()] = http.StatusBadRequest
	httpStatus[errs.FailedValidation.Value()] = http.StatusPreconditionFailed
	httpStatus[errs.NotFound.Value()] = http.StatusNotFound
	httpStatus[errs.MethodNotAllowed.Value()] = http.StatusMethodNotAllowed
	httpStatus[errs.EditConflict.Value()] = http.StatusConflict
	httpStatus[errs.RateLimitExceeded.Value()] = http.StatusTooManyRequests
	httpStatus[errs.AlreadyExists.Value()] = http.StatusConflict
}

func RecoverPanics() web.Middleware {
	mid := func(handler web.Handler) web.Handler {
		hdl := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					e := fmt.Errorf("%v", rec)
					err = errs.NewServerError(errs.Internal, e, errs.InternalMsg)
				}
			}()

			return handler(ctx, w, r)
		}

		return hdl
	}

	return mid
}

func HandleErrors(log *logger.Logger) web.Middleware {
	mid := func(handler web.Handler) web.Handler {
		hdl := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if err := handler(ctx, w, r); err != nil {
				var serverError *errs.ServerError
				var clientError *errs.ClientError

				switch {
				case errors.As(err, &serverError):
					log.Error(ctx, "caught an internal server error during request",
						"ERROR", serverError.LogMsg.Error(),
						"source_err_file", filepath.Base(serverError.FileName),
						"source_err_func", filepath.Base(serverError.FuncName))

					env := web.Envelope{
						"status":  "error",
						"message": serverError.ResMsg.Error(),
					}

					return web.Respond(ctx, w, httpStatus[serverError.Code.Value()], env)

				case errors.As(err, &clientError):
					log.Info(ctx, "caught an error during request",
						"err", clientError.Error(),
						"source_err_file", filepath.Base(clientError.FileName),
						"source_err_func", filepath.Base(clientError.FuncName))

					env := web.Envelope{
						"status":  "fail",
						"message": clientError.ResMsg.Error(),
						"data":    clientError.Data,
					}

					return web.Respond(ctx, w, httpStatus[clientError.Code.Value()], env)

				default:
					log.Error(ctx, "caught an unexpected error during request", "ERROR", err)

					env := web.Envelope{
						"status":  "error",
						"message": errs.InternalMsg,
					}

					return web.Respond(ctx, w, http.StatusInternalServerError, env)
				}
			}

			return nil
		}

		return hdl
	}

	return mid
}

func HandleLogs(log *logger.Logger) web.Middleware {
	mid := func(handler web.Handler) web.Handler {
		hdl := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			tracer := web.GetTracer(ctx)

			log.Info(ctx, "request started",
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"remoteaddr", r.RemoteAddr,
			)

			err := handler(ctx, w, r)

			duration := time.Since(tracer.Now)

			fields := []any{
				"method", r.Method,
				"path", r.URL.RequestURI(),
				"remoteaddr", r.RemoteAddr,
				"statuscode", tracer.StatusCode,
				"since", duration.String(),
			}

			if duration > 500*time.Millisecond {
				log.Warn(ctx, "request completed (slow)", fields...)
			} else {
				log.Info(ctx, "request completed", fields...)
			}

			return err
		}

		return hdl
	}

	return mid
}

func RateLimit(enabled bool, rps float64, burst int) web.Middleware {
	mid := func(handler web.Handler) web.Handler {
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
					return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
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
						return errs.NewClientError(errs.RateLimitExceeded, nil, errs.RateLimitExceededMsg)
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
