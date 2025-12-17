package userapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/agkmw/reddit-clone/internal/app/sdk/errs"
	"github.com/agkmw/reddit-clone/internal/database/userdb"
	"github.com/agkmw/reddit-clone/internal/platform/web"
	"github.com/google/uuid"
)

type api struct {
	db *userdb.Store
}

func newAPI(db *userdb.Store) *api {
	return &api{db: db}
}

func (a *api) RegisterUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := web.Decode(w, r, &input); err != nil {
		return errs.NewClientError(errs.BadRequest, err, errs.BadRequestMsg)
	}

	// TODO: Don't use the db Model to respond back; use app Model;
	user := userdb.User{
		ID:       uuid.New(),
		Username: input.Username,
		Email:    input.Email,
	}

	if err := user.Password.Set(input.Password); err != nil {
		// TODO: Refactor the Error package; current approach of creating
		// errors feels like it needs refactoring...
		return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
	}

	if err := a.db.Create(&user); err != nil {
		switch {
		case errors.Is(err, userdb.ErrUsernameAlreadyExists):
			return errs.NewClientError(errs.AlreadyExists, "user already exists", errs.AlreadyExistsMsg)
		case errors.Is(err, userdb.ErrEmailAlreadyExists):
			return errs.NewClientError(errs.EditConflict, "user already exists", errors.New("user already exists"))
		default:
			return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
		}
	}

	return web.Respond(ctx, w, http.StatusOK, web.Envelope{
		"status": "success",
		"data":   user,
	})
}

func (a *api) GetUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	username := web.ReadParam(r, "username")

	user, err := a.db.GetUserByUsername(username)
	if err != nil {
		switch {
		case errors.Is(err, userdb.ErrRecordNotFound):
			return errs.NewClientError(errs.NotFound, err, errs.NotFoundMsg)
		default:
			return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
		}
	}

	env := web.Envelope{
		"status": "success",
		"data": map[string]any{
			"user": user,
		},
	}

	return web.Respond(ctx, w, http.StatusOK, env)
}

func (a *api) UpdateUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	username := web.ReadParam(r, "username")

	user, err := a.db.GetUserByUsername(username)
	if err != nil {
		switch {
		case errors.Is(err, userdb.ErrRecordNotFound):
			return errs.NewClientError(errs.NotFound, err, errs.NotFoundMsg)
		default:
			return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
		}
	}

	var input struct {
		Username *string `json:"username"`
		Email    *string `json:"email"`
	}

	if err := web.Decode(w, r, &input); err != nil {
		return errs.NewClientError(errs.BadRequest, err, errs.BadRequestMsg)
	}

	if input.Username != nil {
		user.Username = *input.Username
	}

	if input.Email != nil {
		user.Email = *input.Email
	}

	now := time.Now()
	user.LastLogin = &now

	if err := a.db.UpdateUser(user); err != nil {
		switch {
		case errors.Is(err, userdb.ErrUsernameAlreadyExists):
			return errs.NewClientError(errs.AlreadyExists, "user already exists", errs.AlreadyExistsMsg)
		case errors.Is(err, userdb.ErrEmailAlreadyExists):
			return errs.NewClientError(errs.EditConflict, "user already exists", errors.New("user already exists"))
		default:
			return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
		}
	}

	env := web.Envelope{
		"status": "success",
		"data": map[string]any{
			"user": user,
		},
	}

	return web.Respond(ctx, w, http.StatusOK, env)
}

func (a *api) DeleteUserHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	username := web.ReadParam(r, "username")

	if err := a.db.DeleteUser(username); err != nil {
		switch {
		case errors.Is(err, userdb.ErrRecordNotFound):
			return errs.NewClientError(errs.NotFound, err, errs.NotFoundMsg)
		default:
			return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
		}
	}

	return web.Respond(ctx, w, http.StatusOK, web.Envelope{
		"status": "success",
		"data":   "account deleted successfully",
	})
}

func (a *api) ListUsersHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	users, err := a.db.GetUsers()
	if err != nil {
		return errs.NewServerError(errs.Internal, err, errs.InternalMsg)
	}

	env := web.Envelope{
		"status": "success",
		"data": map[string]any{
			"users": users,
		},
	}

	return web.Respond(ctx, w, http.StatusOK, env)
}
