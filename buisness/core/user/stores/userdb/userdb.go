// Package userdb contains user related CRUD functionality.
package userdb

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/mail"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/core/user"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/data/order"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/database/dbarray"
	db "github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/database/pgx"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log *zap.SugaredLogger
	// We are representing here the database connection
	db *sqlx.DB
}

// NewStore constructs the api for data access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :enabled, :department, :date_created, :date_updated)`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, db.ErrDBDuplicatedEntry) {
			return fmt.Errorf("namedexeccontext: %w", user.ErrUniqueEmail)
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"department" = :department,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, db.ErrDBDuplicatedEntry) {
			return user.ErrUniqueEmail
		}
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr user.User) error {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: usr.ID.String(),
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := db.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("namedexeccontext: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter user.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]user.User, error) {
	data := map[string]interface{}{
		"offset":        (pageNumber - 1) * rowsPerPage,
		"rows_per_page": rowsPerPage,
	}

	const q = `
	SELECT
		user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated
	FROM
		users`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	orderByClause, err := orderByClause(orderBy)
	if err != nil {
		return nil, err
	}

	buf.WriteString(orderByClause)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	// We define a slice set to the zero value of type db user
	// we pass the address to the convenience function, once it gets the data
	// we execute over these rows create db user values and append to the slice
	var dbUsrs []dbUser
	if err := db.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &dbUsrs); err != nil {
		return nil, fmt.Errorf("namedqueryslice: %w", err)
	}
	// We convert it back to a core type
	usrs, err := toCoreUserSlice(dbUsrs)
	if err != nil {
		return nil, err
	}

	return usrs, nil
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
	SELECT
		count(1)
	FROM
		users`

	buf := bytes.NewBufferString(q)
	s.applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := db.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
        user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated
	FROM
		users
	WHERE 
		user_id = :user_id`

	var dbUsr dbUser
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return user.User{}, fmt.Errorf("namedquerystruct: %w", user.ErrNotFound)
		}
		return user.User{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	usr, err := toCoreUser(dbUsr)
	if err != nil {
		return user.User{}, err
	}

	return usr, nil
}

// QueryByIDs gets the specified users from the database.
func (s *Store) QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	ids := make([]string, len(userIDs))
	for i, userID := range userIDs {
		ids[i] = userID.String()
	}

	data := struct {
		UserID interface {
			driver.Valuer
			sql.Scanner
		} `db:"user_id"`
	}{
		UserID: dbarray.Array(ids),
	}

	const q = `
	SELECT
        user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated
	FROM
		users
	WHERE
		user_id = ANY(:user_id)`

	var dbUsrs []dbUser
	if err := db.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbUsrs); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("namedquerystruct: %w", err)
	}

	usrs, err := toCoreUserSlice(dbUsrs)
	if err != nil {
		return nil, err
	}

	return usrs, nil
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email.Address,
	}

	const q = `
	SELECT
        user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated
	FROM
		users
	WHERE
		email = :email`

	var dbUsr dbUser
	if err := db.NamedQueryStruct(ctx, s.log, s.db, q, data, &dbUsr); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return user.User{}, fmt.Errorf("namedquerystruct: %w", user.ErrNotFound)
		}
		return user.User{}, fmt.Errorf("namedquerystruct: %w", err)
	}

	usr, err := toCoreUser(dbUsr)
	if err != nil {
		return user.User{}, err
	}

	return usr, nil
}
