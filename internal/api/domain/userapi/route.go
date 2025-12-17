package userapi

import (
	"net/http"

	"github.com/agkmw/reddit-clone/internal/database/userdb"
	"github.com/agkmw/reddit-clone/internal/platform/web"
)

func Routes(app *web.App, db *userdb.Store) {
	api := newAPI(db)

	app.HandlerFunc(http.MethodGet, "/v1", "/users", api.ListUsersHandler)
	app.HandlerFunc(http.MethodPost, "/v1", "/users", api.RegisterUserHandler)
	app.HandlerFunc(http.MethodGet, "/v1", "/users/{username}", api.GetUserHandler)
	app.HandlerFunc(http.MethodPatch, "/v1", "/users/{username}", api.UpdateUserHandler)
	app.HandlerFunc(http.MethodDelete, "/v1", "/users/{username}", api.DeleteUserHandler)
}
