package models

import (
	"database/sql"
	"errors"
	"log"
	"strings"
)

type Permission struct {
	Role       NullString `json:"role" db:"role"`
	Object     NullString `json:"object" db:"object"`
	Permission int        `json:"permission" db:"permission"`
}

type PermissionAPIImpl struct{}

var PermissionAPI PermissionAPIInterface = new(PermissionAPIImpl)

type PermissionAPIInterface interface {
	GetPermission(p Permission) error
	GetPermissionsByRole(role int) ([]Permission, error)
	InsertPermission(p Permission) (Permission, error)
	UpdatePermission(p Permission) (Permission, error)
	DeletePermission(p Permission) error
}

func (a *PermissionAPIImpl) GetPermission(p Permission) error {

	permission := Permission{}
	err := DB.QueryRowx("SELECT * FROM permissions WHERE role = ? AND object = ?", p.Role, p.Object).StructScan(&permission)
	switch {
	case err == sql.ErrNoRows:
		err = errors.New("Project Not Found")
	case err != nil:
		log.Fatal(err)
	default:
		err = nil
	}
	return err
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

func (a *PermissionAPIImpl) InsertPermission(p Permission) (Permission, error) {

	permission := Permission{}
	query, _ := generateSQLStmt(p, "insert", "permissions")
	result, err := DB.NamedExec(query, p)

	if err != nil {
		log.Fatal(err)
		return permission, err
	}
	rowCnt, err := result.RowsAffected()
	if rowCnt > 1 {
		return permission, errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return permission, errors.New("Permission Insert Fail")
	}
	return p, nil
}

func (a *PermissionAPIImpl) UpdatePermission(p Permission) (Permission, error) {

	permission := Permission{}
	query, _ := generateSQLStmt(p, "update", "permission")
	result, err := DB.NamedExec(query, p)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return permission, errors.New("Duplicate entry")
		}
		return permission, err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rowCnt > 1 {
		return permission, errors.New("More Than One Rows Affected") //Transaction rollback?
	} else if rowCnt == 0 {
		return permission, errors.New("No Row Inserted")
	}
	return p, nil
}

func (a *PermissionAPIImpl) DeletePermission(p Permission) error {

	_, err := DB.Exec("DELETE FROM permissions WHERE role = ? AND object = ?", p.Role, p.Object)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
