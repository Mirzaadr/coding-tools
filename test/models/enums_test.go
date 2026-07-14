package models_test

import (
	"testing"

	"github.com/andre/dotnet-architect/internal/models"
)

func TestParsePresentationType_Valid(t *testing.T) {
	cases := []string{"WebApi", "WebApp"}
	for _, raw := range cases {
		if _, err := models.ParsePresentationType(raw); err != nil {
			t.Errorf("expected %q to parse successfully, got error: %v", raw, err)
		}
	}
}

func TestParsePresentationType_Invalid(t *testing.T) {
	if _, err := models.ParsePresentationType("Desktop"); err == nil {
		t.Fatal("expected error for invalid presentation type, got nil")
	}
}

func TestParseDatabaseType_Valid(t *testing.T) {
	cases := []string{"PostgreSQL", "SqlServer"}
	for _, raw := range cases {
		if _, err := models.ParseDatabaseType(raw); err != nil {
			t.Errorf("expected %q to parse successfully, got error: %v", raw, err)
		}
	}
}

func TestParseDatabaseType_Invalid(t *testing.T) {
	if _, err := models.ParseDatabaseType("MongoDB"); err == nil {
		t.Fatal("expected error for invalid database type, got nil")
	}
}
