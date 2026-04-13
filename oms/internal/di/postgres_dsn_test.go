package oms_di

import (
	"strings"
	"testing"
)

func TestPostgresURIWithOMSSearchPath_appendsOptions(t *testing.T) {
	t.Parallel()

	out, err := postgresURIWithOMSSearchPath("postgres://u:p@db.example:5432/shop")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "options=") {
		t.Fatalf("expected options= in %q", out)
	}
	if !strings.Contains(out, "search_path") {
		t.Fatalf("expected search_path in %q", out)
	}
}

func TestPostgresURIWithOMSSearchPath_mergesExistingOptions(t *testing.T) {
	t.Parallel()

	in := "postgresql://u:p@h:5432/db?options=-ctimezone%3DUTC"
	out, err := postgresURIWithOMSSearchPath(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "timezone") {
		t.Fatalf("expected timezone preserved in %q", out)
	}
	if !strings.Contains(out, "search_path") {
		t.Fatalf("expected search_path in %q", out)
	}
}

func TestPostgresURIWithOMSSearchPath_respectsExistingSearchPath(t *testing.T) {
	t.Parallel()

	in := "postgres://u:p@h:5432/db?options=-csearch_path%3Ddelivery%2Cpublic"
	out, err := postgresURIWithOMSSearchPath(in)
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("expected unchanged URL, got %q want %q", out, in)
	}
}

func TestPostgresURIWithOMSSearchPath_nonPostgresUnchanged(t *testing.T) {
	t.Parallel()

	in := "sqlite://./local.db"
	out, err := postgresURIWithOMSSearchPath(in)
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("got %q want %q", out, in)
	}
}
