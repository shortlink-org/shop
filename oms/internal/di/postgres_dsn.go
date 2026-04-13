package oms_di

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/shortlink-org/go-sdk/config"
)

const omsPostgresSearchPathOptions = "-csearch_path=oms,public"

// postgresURIWithOMSSearchPath appends a session `search_path` so unqualified names (including
// golang-migrate's schema_migrations_* table) resolve to schema `oms` first.
//
// PostgreSQL 15+ no longer grants CREATE on schema public to all roles. Application roles
// created for OMS typically own schema `oms` only.
//
// If the URL already sets search_path via the `options` query parameter, it is left unchanged.
func postgresURIWithOMSSearchPath(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", err
	}

	switch strings.ToLower(u.Scheme) {
	case "postgres", "postgresql":
	default:
		return databaseURL, nil
	}

	q := u.Query()
	opts := q.Get("options")
	if strings.Contains(strings.ToLower(opts), "search_path") {
		return databaseURL, nil
	}

	if opts == "" {
		q.Set("options", omsPostgresSearchPathOptions)
	} else {
		q.Set("options", opts+" "+omsPostgresSearchPathOptions)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// provideOMSConfig loads configuration and normalizes the Postgres DSN when OMS uses PostgreSQL.
func provideOMSConfig() (*config.Config, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	if !strings.EqualFold(cfg.GetString("STORE_TYPE"), "postgres") {
		return cfg, nil
	}

	raw := cfg.GetString("STORE_POSTGRES_URI")
	if raw == "" {
		return cfg, nil
	}

	out, err := postgresURIWithOMSSearchPath(raw)
	if err != nil {
		return nil, fmt.Errorf("normalize STORE_POSTGRES_URI: %w", err)
	}

	cfg.Set("STORE_POSTGRES_URI", out)

	return cfg, nil
}
