// Package usergrp maintains the group of handlers for user access.
package usrgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/core/user"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/validate"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/auth"
	v1 "github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1"
	paging "github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/web/v1/paging"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/foundation/web"
	"github.com/golang-jwt/jwt/v4"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	user *user.Core
	auth *auth.Auth
}

// New constructs a handlers for route access.
func New(user *user.Core) *Handlers {
	return &Handlers{
		user: user,
	}
}

// Create adds a new user to the system.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var app AppNewUser
	// we decode using the json unmarshaler then validate
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	// If we reach here then we know that the data is trusted
	// We convert it to a core model
	nc, err := toCoreNewUser(app)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	usr, err := h.user.Create(ctx, nc)
	if err != nil {
		if errors.Is(err, user.ErrUniqueEmail) {
			// we return  a trusted error
			return v1.NewRequestError(err, http.StatusConflict)
		}
		return fmt.Errorf("create: usr[%+v]: %w", usr, err)
	}

	// toAppUser() convert the buisness model to the app
	return web.Respond(ctx, w, toAppUser(usr), http.StatusCreated)
}

// // Update updates a user in the system.
// func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

// 	var app AppUpdateUser
// 	if err := web.Decode(r, &app); err != nil {
// 		return err
// 	}
// 	// We are hitting the DB so we can check userID
// 	userID := auth.GetUserID(ctx)

// 	usr, err := h.user.QueryByID(ctx, userID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, user.ErrNotFound):
// 			return v1.NewRequestError(err, http.StatusNotFound)
// 		default:
// 			return fmt.Errorf("querybyid: userID[%s]: %w", userID, err)
// 		}
// 	}

// 	uu, err := toCoreUpdateUser(app)
// 	if err != nil {
// 		return return v1.NewRequestError(err, http.StatusBadRequest)
// 	}

// 	usr, err = h.user.Update(ctx, usr, uu)
// 	if err != nil {
// 		return fmt.Errorf("update: userID[%s] uu[%+v]: %w", userID, uu, err)
// 	}

// 	return web.Respond(ctx, w, toAppUser(usr), http.StatusOK)
// }

// // Delete removes a user from the system.
// func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
// 	userID := auth.GetUserID(ctx)

// 	h, err := h.executeUnderTransaction(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	usr, err := h.user.QueryByID(ctx, userID)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, user.ErrNotFound):
// 			return web.Respond(ctx, w, nil, http.StatusNoContent)
// 		default:
// 			return fmt.Errorf("querybyid: userID[%s]: %w", userID, err)
// 		}
// 	}

// 	if err := h.user.Delete(ctx, usr); err != nil {
// 		return fmt.Errorf("delete: userID[%s]: %w", userID, err)
// 	}

// 	return web.Respond(ctx, w, nil, http.StatusNoContent)
// }

// Query returns a list of users with paging.
func (h *Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.Parse(r)
	if err != nil {
		return err
	}

	filter, err := parseFilter(r)
	if err != nil {
		return err
	}

	orderBy, err := parseOrder(r)
	if err != nil {
		return err
	}

	users, err := h.user.Query(ctx, filter, orderBy, page.Number, page.RowsPerPage)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	total, err := h.user.Count(ctx, filter)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(toAppUsers(users), total, page.Number, page.RowsPerPage), http.StatusOK)
}

// // QueryByID returns a user by its ID.
// func (h *Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
// 	id := auth.GetUserID(ctx)

// 	usr, err := h.user.QueryByID(ctx, id)
// 	if err != nil {
// 		switch {
// 		case errors.Is(err, user.ErrNotFound):
// 			return response.NewError(err, http.StatusNotFound)
// 		default:
// 			return fmt.Errorf("querybyid: id[%s]: %w", id, err)
// 		}
// 	}

// 	return web.Respond(ctx, w, toAppUser(usr), http.StatusOK)
// }

// Token provides an API token for the authenticated user.
func (h *Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	kid := web.Param(r, "kid")
	if kid == "" {
		return validate.NewFieldsError("kid", errors.New("missing kid"))
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		return auth.NewAuthError("must provide email and password in Basic auth")
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return auth.NewAuthError("invalid email format")
	}

	usr, err := h.user.Authenticate(ctx, *addr, pass)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return v1.NewRequestError(err, http.StatusNotFound)
		case errors.Is(err, user.ErrAuthenticationFailure):
			return auth.NewAuthError(err.Error())
		default:
			return fmt.Errorf("authenticate: %w", err)
		}
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID.String(),
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: usr.Roles,
	}

	token, err := h.auth.GenerateToken(kid, claims)
	if err != nil {
		return fmt.Errorf("generatetoken: %w", err)
	}

	return web.Respond(ctx, w, toToken(token), http.StatusOK)
}
