package portaldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var ErrUnavailable = errors.New("portal database is unavailable")

type Config struct {
	Enabled     bool
	Driver      string
	DSN         string
	AutoMigrate bool
}

type Client interface {
	Healthy(ctx context.Context) error
	Enabled() bool
	DB() *sql.DB
	Driver() string
	EnsureSchema(ctx context.Context) error
}

type FakeClient struct {
	config Config
}

type SQLClient struct {
	config Config
	db     *sql.DB
	err    error
}

func New(cfg Config) Client {
	if !cfg.Enabled || strings.TrimSpace(cfg.DSN) == "" {
		return &FakeClient{config: cfg}
	}

	driver := normalizeDriver(cfg.Driver, cfg.DSN)
	db, err := sql.Open(driver, cfg.DSN)
	if err != nil {
		return &SQLClient{config: cfg, err: err}
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(8)

	client := &SQLClient{
		config: cfg,
		db:     db,
	}
	client.config.Driver = driver
	if cfg.AutoMigrate {
		client.err = client.EnsureSchema(context.Background())
	}
	return client
}

func NewFromDB(cfg Config, db *sql.DB) Client {
	return &SQLClient{
		config: Config{
			Enabled:     true,
			Driver:      normalizeDriver(cfg.Driver, cfg.DSN),
			DSN:         cfg.DSN,
			AutoMigrate: cfg.AutoMigrate,
		},
		db: db,
	}
}

func (c *FakeClient) Healthy(ctx context.Context) error { return ErrUnavailable }
func (c *FakeClient) Enabled() bool                     { return false }
func (c *FakeClient) DB() *sql.DB                       { return nil }
func (c *FakeClient) Driver() string                    { return normalizeDriver(c.config.Driver, c.config.DSN) }
func (c *FakeClient) EnsureSchema(ctx context.Context) error {
	return ErrUnavailable
}

func (c *SQLClient) Healthy(ctx context.Context) error {
	if c.err != nil {
		return c.err
	}
	if c.db == nil {
		return ErrUnavailable
	}
	return c.db.PingContext(ctx)
}

func (c *SQLClient) Enabled() bool {
	return c.db != nil && (c.config.Enabled || strings.TrimSpace(c.config.DSN) != "")
}

func (c *SQLClient) DB() *sql.DB {
	return c.db
}

func (c *SQLClient) Driver() string {
	return c.config.Driver
}

func (c *SQLClient) EnsureSchema(ctx context.Context) error {
	if c.db == nil {
		return ErrUnavailable
	}

	ddl := []string{
		`CREATE TABLE IF NOT EXISTS portal_departments (
			department_id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			parent_department_id VARCHAR(64) NULL,
			admin_consumer_name VARCHAR(255) NULL,
			deleted TINYINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_users (
			consumer_name VARCHAR(255) PRIMARY KEY,
			display_name VARCHAR(255) NULL,
			email VARCHAR(255) NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'pending',
			user_level VARCHAR(32) NOT NULL DEFAULT 'normal',
			source VARCHAR(64) NOT NULL DEFAULT 'console',
			password_hash VARCHAR(255) NULL,
			temp_password VARCHAR(255) NULL,
			department_id VARCHAR(64) NULL,
			parent_consumer_name VARCHAR(255) NULL,
			is_department_admin TINYINT NOT NULL DEFAULT 0,
			last_login_at TIMESTAMP NULL,
			deleted TINYINT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_invite_code (
			invite_code VARCHAR(64) PRIMARY KEY,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			expires_at TIMESTAMP NULL,
			used_by_consumer VARCHAR(255) NULL,
			used_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_asset_grant (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			asset_type VARCHAR(64) NOT NULL,
			asset_id VARCHAR(255) NOT NULL,
			subject_type VARCHAR(64) NOT NULL,
			subject_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_model_asset (
			asset_id VARCHAR(255) PRIMARY KEY,
			canonical_name VARCHAR(255) NOT NULL,
			display_name VARCHAR(255) NOT NULL,
			intro TEXT NULL,
			tags_json TEXT NULL,
			capabilities_json TEXT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_model_binding (
			binding_id VARCHAR(255) PRIMARY KEY,
			asset_id VARCHAR(255) NOT NULL,
			model_id VARCHAR(255) NOT NULL,
			provider_name VARCHAR(255) NOT NULL,
			target_model VARCHAR(255) NOT NULL,
			protocol VARCHAR(64) NOT NULL DEFAULT 'openai/v1',
			endpoint TEXT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'draft',
			published_at TIMESTAMP NULL,
			unpublished_at TIMESTAMP NULL,
			pricing_json TEXT NULL,
			limits_json TEXT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_model_binding_price_version (
			version_id BIGINT PRIMARY KEY AUTO_INCREMENT,
			asset_id VARCHAR(255) NOT NULL,
			binding_id VARCHAR(255) NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			active TINYINT NOT NULL DEFAULT 0,
			effective_from TIMESTAMP NULL,
			effective_to TIMESTAMP NULL,
			pricing_json TEXT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_agent_catalog (
			agent_id VARCHAR(255) PRIMARY KEY,
			canonical_name VARCHAR(255) NOT NULL,
			display_name VARCHAR(255) NOT NULL,
			intro TEXT NULL,
			description TEXT NULL,
			icon_url TEXT NULL,
			tags_json TEXT NULL,
			mcp_server_name VARCHAR(255) NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'draft',
			tool_count INT NOT NULL DEFAULT 0,
			transport_types_json TEXT NULL,
			resource_summary TEXT NULL,
			prompt_summary TEXT NULL,
			published_at TIMESTAMP NULL,
			unpublished_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_ai_sensitive_detect_rule (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			pattern TEXT NOT NULL,
			match_type VARCHAR(32) NOT NULL,
			description TEXT NULL,
			priority INT NOT NULL DEFAULT 0,
			enabled TINYINT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_ai_sensitive_replace_rule (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			pattern TEXT NOT NULL,
			replace_type VARCHAR(32) NOT NULL,
			replace_value TEXT NULL,
			restore TINYINT NOT NULL DEFAULT 0,
			description TEXT NULL,
			priority INT NOT NULL DEFAULT 0,
			enabled TINYINT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_ai_sensitive_system_config (
			config_key VARCHAR(64) PRIMARY KEY,
			system_deny_enabled TINYINT NOT NULL DEFAULT 0,
			dictionary_text LONGTEXT NULL,
			updated_by VARCHAR(255) NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS portal_ai_sensitive_block_audit (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			request_id VARCHAR(255) NULL,
			route_name VARCHAR(255) NULL,
			consumer_name VARCHAR(255) NULL,
			display_name VARCHAR(255) NULL,
			blocked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			blocked_by VARCHAR(255) NULL,
			request_phase VARCHAR(64) NULL,
			blocked_reason_json LONGTEXT NULL,
			match_type VARCHAR(32) NULL,
			matched_rule TEXT NULL,
			matched_excerpt TEXT NULL,
			provider_id BIGINT NULL,
			cost_usd VARCHAR(64) NULL
		)`,
		`CREATE TABLE IF NOT EXISTS job_run_record (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			job_name VARCHAR(255) NOT NULL,
			trigger_source VARCHAR(64) NOT NULL,
			trigger_id VARCHAR(255) NOT NULL,
			status VARCHAR(32) NOT NULL,
			idempotency_key VARCHAR(255) NULL,
			target_version VARCHAR(255) NULL,
			message TEXT NULL,
			error_text LONGTEXT NULL,
			started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			finished_at TIMESTAMP NULL,
			duration_ms BIGINT NULL
		)`,
	}

	for _, statement := range ddl {
		if _, err := c.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDriver(driver, dsn string) string {
	normalized := strings.ToLower(strings.TrimSpace(driver))
	if normalized != "" {
		return normalized
	}
	return "mysql"
}

func MustDB(client Client) (*sql.DB, error) {
	if client == nil || !client.Enabled() || client.DB() == nil {
		return nil, ErrUnavailable
	}
	return client.DB(), nil
}

func WrapExecError(action string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", action, err)
}
