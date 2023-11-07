package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresInput represents a sample postgres input source
//
// This source will:
//
//  1. Create a function which notifies a channel with a json payload representing an operation
//  2. Add a trigger to every table in a database to call that function on Creat, Update, and Deletes
//  3. Listen to the channel created in step 1
//
// The operations passed by the database can then be passed to a Process
type PostgresInput struct {
	conn     *sqlx.DB
	listener *pq.Listener
	config   InputConfig
}

type postgresTriggerResult struct {
	Table     string `json:"tbl"`
	ID        any    `json:"id"`
	Operation string `json:"op"`
}

// NewPostgresInput accepts an InputConfig and returns a PostgresInput,
// which implements the orchestrator.Input interface
//
// The InputConfig.ConnectionString argument can be a DSN, or a postgres
// URL
func NewPostgresInput(ic InputConfig) (p PostgresInput, err error) {
	p.config = ic

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	url := ic.ConnectionString
	if strings.HasPrefix(url, "postgres://") || strings.HasPrefix(url, "postgresql://") {
		url, err = pq.ParseURL(ic.ConnectionString)
		if err != nil {
			return
		}
	}

	p.conn, err = sqlx.ConnectContext(ctx, "postgres", url)
	if err != nil {
		return
	}

	p.listener = pq.NewListener(ic.ConnectionString, time.Second, time.Second*10, func(event pq.ListenerEventType, err error) {
		if err != nil {
			panic(err)
		}
	})

	return
}

// ID returns the ID for this Input
func (p PostgresInput) ID() string {
	return p.config.ID()
}

// Handle will configure a database for notification, and then listen to those
// notifications
func (p PostgresInput) Handle(ctx context.Context, c chan Event) (err error) {
	err = p.createTriggers()
	if err != nil {
		return
	}

	err = p.listener.Listen(p.ID())
	if err != nil {
		return err
	}

	for n := range p.listener.NotificationChannel() {
		input := new(postgresTriggerResult)

		err = json.Unmarshal([]byte(n.Extra), input)
		if err != nil {
			return
		}

		eo := OperationUnknown
		switch input.Operation {
		case "INSERT":
			eo = OperationCreate
		}

		c <- Event{
			Location:  input.Table,
			Operation: eo,
			ID:        fmt.Sprintf("%v", input.ID),
			Trigger:   p.ID(),
		}
	}

	return
}

// createTriggers will connect to the database and configure triggers and
// notifies ahead of processing
func (p PostgresInput) createTriggers() (err error) {
	tf := p.triggerFunc()
	tx := p.conn.MustBegin()

	_, err = tx.Exec(tf)
	if err != nil {
		return
	}

	tables := make([]string, 0)
	err = tx.Select(&tables, "SELECT tablename FROM pg_catalog.pg_tables where schemaname = 'public';")
	if err != nil {
		return
	}

	for _, table := range tables {
		_, err = tx.Exec(p.addTrigger(table))
		if err != nil {
			return
		}
	}

	return tx.Commit()
}

func (p PostgresInput) triggerFunc() string {
	return fmt.Sprintf(`CREATE OR REPLACE FUNCTION process_record_%[1]s() RETURNS TRIGGER as $process_record_%[1]s$
BEGIN
    PERFORM pg_notify('%[1]s', json_build_object('tbl', TG_TABLE_NAME, 'id', COALESCE(NEW.id, 0), 'op', TG_OP)::Text);
    RETURN NEW;
END;
$process_record_%[1]s$ LANGUAGE plpgsql;`, p.ID())
}

func (p PostgresInput) addTrigger(table string) string {
	return fmt.Sprintf(`CREATE OR REPLACE TRIGGER %[1]s_%[2]s_trigger
AFTER INSERT OR UPDATE OR DELETE ON %[1]s FOR EACH ROW
EXECUTE PROCEDURE process_record_%[2]s();`, table, p.ID())
}
