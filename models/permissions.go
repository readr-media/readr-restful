package models

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"database/sql"
	"github.com/go-sql-driver/mysql"
)

type Permission struct {
	Role       int        `json:"role" db:"role"`
	Object     NullString `json:"object" db:"object"`
	Permission int        `json:"permission" db:"permission"`
}

type PermissionAPIImpl struct{}

var PermissionAPI PermissionAPIInterface = new(PermissionAPIImpl)

type PermissionAPIInterface interface {
	GetPermissions(ps []Permission) ([]Permission, error)
	GetPermissionsByRole(role int) ([]Permission, error)
	GetPermissionsAll() ([]Permission, error)
	InsertPermissions(ps []Permission) error
	DeletePermissions(ps []Permission) error
}

func (a *PermissionAPIImpl) GetPermissions(ps []Permission) ([]Permission, error) {

	var queryHead = "SELECT * FROM permissions WHERE "
	var queryBody []string
	for _, p := range ps {
		queryBody = append(queryBody, fmt.Sprintf("(role = \"%d\" AND object = \"%s\")", p.Role, p.Object.String))
	}
	query := queryHead + strings.Join(queryBody[:], " OR ") + ";"

	rows, err := DB.Queryx(query)
	switch {
	case err != nil:
		log.Fatal(err)
	default:
		err = nil
	}

	permissions := []Permission{}
	for rows.Next() {
		var p Permission
		err = rows.StructScan(&p)
		permissions = append(permissions, p)
	}
	return permissions, err
}

func (a *PermissionAPIImpl) GetPermissionsByRole(role int) ([]Permission, error) {

	permissions := []Permission{}
	rows, err := DB.Queryx("SELECT * FROM permissions WHERE role = ?", role)
	switch {
	case err == sql.ErrNoRows:
		permissions = nil
	case err != nil:
		log.Fatal(err)
		permissions = nil
	default:
		err = nil
	}

	for rows.Next() {
		var p Permission
		err = rows.StructScan(&p)
		permissions = append(permissions, p)
	}
	return permissions, err
}

func (a *PermissionAPIImpl) GetPermissionsAll() ([]Permission, error) {

	permissions := []Permission{}
	rows, err := DB.Queryx("SELECT * FROM permissions")
	switch {
	case err == sql.ErrNoRows:
		permissions = nil
	case err != nil:
		log.Fatal(err)
		permissions = nil
	default:
		err = nil
	}
	for rows.Next() {
		var p Permission
		err = rows.StructScan(&p)
		permissions = append(permissions, p)
	}
	return permissions, err
}

func (a *PermissionAPIImpl) InsertPermissions(ps []Permission) error {

	var queryHead = "INSERT INTO permissions (role, object, permission) VALUE "
	var queryBody []string
	for _, p := range ps {
		queryBody = append(queryBody, fmt.Sprintf("(\"%d\",\"%s\",\"%s\")", p.Role, p.Object.String, "1"))
	}
	query := queryHead + strings.Join(queryBody[:], ",") + ";"

	_, err := DB.Exec(query)

	if err != nil {
		switch err.(*mysql.MySQLError).Number {
		case 1062:
			return errors.New("Duplicate Entry")
		default:
			return err
		}
	}

	return nil
}

func (a *PermissionAPIImpl) DeletePermissions(ps []Permission) error {

	var queryHead = "DELETE FROM permissions WHERE "
	var queryBody []string
	for _, p := range ps {
		queryBody = append(queryBody, fmt.Sprintf("(role = \"%d\" AND object = \"%s\")", p.Role, p.Object.String))
	}
	query := queryHead + strings.Join(queryBody[:], " OR ") + ";"

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
