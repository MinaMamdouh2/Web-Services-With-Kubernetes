package userdb

import (
	"database/sql"
	"fmt"
	"net/mail"
	"time"

	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/core/user"
	"github.com/MinaMamdouh2/Web-Services-With-Kubernetes/buisness/sys/database/dbarray"
	"github.com/google/uuid"
)

// dbUser represent the structure we need for moving data
// between the app and the database.
// We use this convention dbUser and thid type is not exported
// it lives inside this package only, it is for internal use
// and it uses the tagging system, since this what sqlx use
type dbUser struct {
	ID           uuid.UUID      `db:"user_id"`
	Name         string         `db:"name"`
	Email        string         `db:"email"`
	Roles        dbarray.String `db:"roles"`
	PasswordHash []byte         `db:"password_hash"`
	Enabled      bool           `db:"enabled"`
	Department   sql.NullString `db:"department"`
	DateCreated  time.Time      `db:"date_created"`
	DateUpdated  time.Time      `db:"date_updated"`
}

// FROM a buisness core user model,
// we marshel that to a DB user model
func toDBUser(usr user.User) dbUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return dbUser{
		ID:           usr.ID,
		Name:         usr.Name,
		Email:        usr.Email.Address,
		Roles:        roles,
		PasswordHash: usr.PasswordHash,
		Department: sql.NullString{
			String: usr.Department,
			Valid:  usr.Department != "",
		},
		Enabled:     usr.Enabled,
		DateCreated: usr.DateCreated.UTC(),
		DateUpdated: usr.DateUpdated.UTC(),
	}
}

// We got data out the database, so we get it back as a dbUser
func toCoreUser(dbUsr dbUser) (user.User, error) {
	addr := mail.Address{
		Address: dbUsr.Email,
	}

	roles := make([]user.Role, len(dbUsr.Roles))
	for i, value := range dbUsr.Roles {
		var err error
		roles[i], err = user.ParseRole(value)
		if err != nil {
			return user.User{}, fmt.Errorf("parse role: %w", err)
		}
	}

	usr := user.User{
		ID:           dbUsr.ID,
		Name:         dbUsr.Name,
		Email:        addr,
		Roles:        roles,
		PasswordHash: dbUsr.PasswordHash,
		Enabled:      dbUsr.Enabled,
		Department:   dbUsr.Department.String,
		DateCreated:  dbUsr.DateCreated.In(time.Local),
		DateUpdated:  dbUsr.DateUpdated.In(time.Local),
	}

	return usr, nil
}

func toCoreUserSlice(dbUsers []dbUser) ([]user.User, error) {
	usrs := make([]user.User, len(dbUsers))
	for i, dbUsr := range dbUsers {
		var err error
		usrs[i], err = toCoreUser(dbUsr)
		if err != nil {
			return nil, err
		}
	}
	return usrs, nil
}
