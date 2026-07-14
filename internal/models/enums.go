package models

import "fmt"

// PresentationType represents the presentation layer flavor to generate.
type PresentationType string

const (
	PresentationWebApi PresentationType = "WebApi"
	PresentationWebApp PresentationType = "WebApp"
)

// ParsePresentationType validates and converts a raw string into a PresentationType.
func ParsePresentationType(raw string) (PresentationType, error) {
	switch PresentationType(raw) {
	case PresentationWebApi:
		return PresentationWebApi, nil
	case PresentationWebApp:
		return PresentationWebApp, nil
	default:
		return "", fmt.Errorf("invalid presentation type %q (allowed: %s, %s)", raw, PresentationWebApi, PresentationWebApp)
	}
}

// DatabaseType represents the database provider to configure.
type DatabaseType string

const (
	DatabasePostgreSQL DatabaseType = "PostgreSQL"
	DatabaseSqlServer  DatabaseType = "SqlServer"
)

// ParseDatabaseType validates and converts a raw string into a DatabaseType.
func ParseDatabaseType(raw string) (DatabaseType, error) {
	switch DatabaseType(raw) {
	case DatabasePostgreSQL:
		return DatabasePostgreSQL, nil
	case DatabaseSqlServer:
		return DatabaseSqlServer, nil
	default:
		return "", fmt.Errorf("invalid database type %q (allowed: %s, %s)", raw, DatabasePostgreSQL, DatabaseSqlServer)
	}
}
