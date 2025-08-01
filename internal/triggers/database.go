package triggers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/logimos/conduktr/internal/engine"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
	"go.uber.org/zap"
)

// DatabaseTrigger implements database change detection
type DatabaseTrigger struct {
	db     *sql.DB
	engine *engine.Engine
	logger *zap.Logger
	config DatabaseConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// DatabaseConfig holds database connection and monitoring configuration
type DatabaseConfig struct {
	Driver       string        `yaml:"driver"`        // postgres, mysql
	DSN          string        `yaml:"dsn"`           // connection string
	PollInterval time.Duration `yaml:"poll_interval"` // polling interval
	Tables       []TableConfig `yaml:"tables"`        // tables to monitor
	UseWAL       bool          `yaml:"use_wal"`       // use WAL for PostgreSQL
}

// TableConfig defines table monitoring configuration
type TableConfig struct {
	Name         string   `yaml:"name"`
	Events       []string `yaml:"events"`        // INSERT, UPDATE, DELETE
	Columns      []string `yaml:"columns"`       // columns to monitor (for updates)
	Condition    string   `yaml:"condition"`     // WHERE condition
	TimestampCol string   `yaml:"timestamp_col"` // timestamp column for polling
}

// ChangeRecord represents a database change
type ChangeRecord struct {
	Table      string                 `json:"table"`
	Operation  string                 `json:"operation"`
	Timestamp  time.Time              `json:"timestamp"`
	OldData    map[string]interface{} `json:"old_data,omitempty"`
	NewData    map[string]interface{} `json:"new_data,omitempty"`
	PrimaryKey map[string]interface{} `json:"primary_key"`
}

// NewDatabaseTrigger creates a new database trigger
func NewDatabaseTrigger(config DatabaseConfig, engine *engine.Engine, logger *zap.Logger) *DatabaseTrigger {
	ctx, cancel := context.WithCancel(context.Background())

	return &DatabaseTrigger{
		engine: engine,
		logger: logger,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins monitoring database changes
func (d *DatabaseTrigger) Start() error {
	var err error
	d.db, err = sql.Open(d.config.Driver, d.config.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := d.db.PingContext(d.ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.logger.Info("Database trigger started",
		zap.String("driver", d.config.Driver),
		zap.Int("tables", len(d.config.Tables)))

	// Start monitoring based on driver capabilities
	switch d.config.Driver {
	case "postgres":
		if d.config.UseWAL {
			go d.listenPostgreSQLWAL()
		} else {
			go d.pollChanges()
		}
	case "mysql":
		go d.pollChanges() // MySQL polling
	default:
		go d.pollChanges() // Generic polling
	}

	return nil
}

// Stop stops the database trigger
func (d *DatabaseTrigger) Stop() error {
	d.cancel()
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// pollChanges implements polling-based change detection
func (d *DatabaseTrigger) pollChanges() {
	ticker := time.NewTicker(d.config.PollInterval)
	defer ticker.Stop()

	lastChecked := make(map[string]time.Time)

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			for _, table := range d.config.Tables {
				d.pollTable(table, lastChecked[table.Name])
				lastChecked[table.Name] = time.Now()
			}
		}
	}
}

// pollTable polls a specific table for changes
func (d *DatabaseTrigger) pollTable(table TableConfig, lastCheck time.Time) {
	if table.TimestampCol == "" {
		d.logger.Warn("No timestamp column configured for polling", zap.String("table", table.Name))
		return
	}

	query := d.buildPollQuery(table, lastCheck)

	rows, err := d.db.QueryContext(d.ctx, query)
	if err != nil {
		d.logger.Error("Failed to poll table",
			zap.String("table", table.Name),
			zap.Error(err))
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		d.logger.Error("Failed to get columns", zap.Error(err))
		return
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			d.logger.Error("Failed to scan row", zap.Error(err))
			continue
		}

		// Convert to map
		rowData := make(map[string]interface{})
		for i, col := range columns {
			rowData[col] = values[i]
		}

		// Create change record
		change := ChangeRecord{
			Table:     table.Name,
			Operation: "UPDATE", // Polling only detects updates/inserts
			Timestamp: time.Now(),
			NewData:   rowData,
		}

		d.handleChange(change)
	}
}

// buildPollQuery builds a query for polling table changes
func (d *DatabaseTrigger) buildPollQuery(table TableConfig, lastCheck time.Time) string {
	query := fmt.Sprintf("SELECT * FROM %s", table.Name)

	conditions := []string{}

	if !lastCheck.IsZero() {
		conditions = append(conditions, fmt.Sprintf("%s > '%s'", table.TimestampCol, lastCheck.Format(time.RFC3339)))
	}

	if table.Condition != "" {
		conditions = append(conditions, table.Condition)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY %s DESC LIMIT 1000", table.TimestampCol)

	return query
}

// listenPostgreSQLWAL implements PostgreSQL WAL-based change detection
func (d *DatabaseTrigger) listenPostgreSQLWAL() {
	// This is a simplified implementation
	// In a production system, you'd use logical replication slots
	d.logger.Info("Starting PostgreSQL WAL listener")

	// For now, fall back to polling
	// TODO: Implement proper logical replication
	d.pollChanges()
}

// handleChange processes a database change
func (d *DatabaseTrigger) handleChange(change ChangeRecord) {
	d.logger.Info("Database change detected",
		zap.String("table", change.Table),
		zap.String("operation", change.Operation))

	// Determine event type
	eventType := fmt.Sprintf("db.%s.%s",
		strings.ToLower(change.Table),
		strings.ToLower(change.Operation))

	// Create context with database-specific metadata
	context := map[string]interface{}{
		"trigger_type": "database",
		"table":        change.Table,
		"operation":    change.Operation,
		"timestamp":    change.Timestamp.Unix(),
		"event_type":   eventType,
	}

	// Add change data to context
	if change.NewData != nil {
		context["new_data"] = change.NewData
	}
	if change.OldData != nil {
		context["old_data"] = change.OldData
	}
	if change.PrimaryKey != nil {
		context["primary_key"] = change.PrimaryKey
	}

	// Execute workflow asynchronously
	go executeWorkflow(d.ctx, d.engine, d.logger, eventType, context)
}

// CreateTrigger creates database triggers for change detection (PostgreSQL)
func (d *DatabaseTrigger) CreateTrigger(tableName string) error {
	if d.config.Driver != "postgres" {
		return fmt.Errorf("database triggers only supported for PostgreSQL")
	}

	// Create trigger function
	functionSQL := fmt.Sprintf(`
                CREATE OR REPLACE FUNCTION reactor_notify_%s()
                RETURNS TRIGGER AS $$
                BEGIN
                        IF TG_OP = 'DELETE' THEN
                                PERFORM pg_notify('reactor_changes', json_build_object(
                                        'table', TG_TABLE_NAME,
                                        'operation', TG_OP,
                                        'old_data', row_to_json(OLD)
                                )::text);
                                RETURN OLD;
                        ELSE
                                PERFORM pg_notify('reactor_changes', json_build_object(
                                        'table', TG_TABLE_NAME,
                                        'operation', TG_OP,
                                        'new_data', row_to_json(NEW),
                                        'old_data', CASE WHEN TG_OP = 'UPDATE' THEN row_to_json(OLD) ELSE NULL END
                                )::text);
                                RETURN NEW;
                        END IF;
                END;
                $$ LANGUAGE plpgsql;
        `, tableName)

	if _, err := d.db.ExecContext(d.ctx, functionSQL); err != nil {
		return fmt.Errorf("failed to create trigger function: %w", err)
	}

	// Create trigger
	triggerSQL := fmt.Sprintf(`
                DROP TRIGGER IF EXISTS reactor_trigger_%s ON %s;
                CREATE TRIGGER reactor_trigger_%s
                AFTER INSERT OR UPDATE OR DELETE ON %s
                FOR EACH ROW EXECUTE FUNCTION reactor_notify_%s();
        `, tableName, tableName, tableName, tableName, tableName)

	if _, err := d.db.ExecContext(d.ctx, triggerSQL); err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	d.logger.Info("Created database trigger", zap.String("table", tableName))
	return nil
}
