package usrgrp

import (
	"errors"
	"net/http"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/core/user"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/data/order"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/validate"
)

func parseOrder(r *http.Request) (order.By, error) {
	const (
		orderByID      = "user_id"
		orderByName    = "name"
		orderByEmail   = "email"
		orderByRoles   = "roles"
		orderByEnabled = "enabled"
	)

	var orderByFields = map[string]string{
		orderByID:      user.OrderByID,
		orderByName:    user.OrderByName,
		orderByEmail:   user.OrderByEmail,
		orderByRoles:   user.OrderByRoles,
		orderByEnabled: user.OrderByEnabled,
	}

	orderBy, err := order.Parse(r, order.NewBy(orderByID, order.ASC))
	if err != nil {
		return order.By{}, err
	}

	if _, exists := orderByFields[orderBy.Field]; !exists {
		return order.By{}, validate.NewFieldsError(orderBy.Field, errors.New("order field does not exist"))
	}

	orderBy.Field = orderByFields[orderBy.Field]

	return orderBy, nil
}
