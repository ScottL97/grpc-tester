package tester

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strings"
)

type CockroachDB struct {
	conn   *pgx.Conn
	config *CockroachDBConfig
}

type KeyPair struct {
	key   string
	value interface{}
}

func NewCockroachDB(c *CockroachDBConfig) *CockroachDB {
	crdbConfig, err := pgx.ParseConfig(buildCockroachDBConnectionString(c))
	if err != nil {
		panic(err)
	}
	conn, err := pgx.ConnectConfig(context.Background(), crdbConfig)
	if err != nil {
		panic(err)
	}

	return &CockroachDB{
		conn:   conn,
		config: c,
	}
}

func buildCockroachDBConnectionString(c *CockroachDBConfig) string {
	// 参考 https://www.cockroachlabs.com/docs/stable/connection-parameters.html
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?application_name=%s&sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.ApplicationName, c.SSLMode)
}

func buildQuery(tableName string, attributes []string, conditions []KeyPair) string {
	sb := strings.Builder{}
	sb.WriteString("SELECT")
	if len(attributes) == 0 {
		sb.WriteString(" *")
	} else {
		for i := 0; i < len(attributes); i++ {
			sb.WriteString(" ")
			sb.WriteString(attributes[i])
			if i != len(attributes)-1 {
				sb.WriteString(",")
			}
		}
	}
	sb.WriteString(" FROM ")
	sb.WriteString(tableName)
	sb.WriteString(" WHERE ")
	for i := 0; i < len(conditions); i++ {
		sb.WriteString(conditions[i].key)
		sb.WriteString(" = $")
		sb.WriteString(fmt.Sprintf("%d", i+1))
		if i != len(conditions)-1 {
			sb.WriteString(" AND ")
		}
	}
	return sb.String()
}

func buildInsert(tableName string, keyPairs []KeyPair) string {
	sb := strings.Builder{}
	sb.WriteString("INSERT INTO ")
	sb.WriteString(tableName)
	sb.WriteString(" (")
	for i := 0; i < len(keyPairs); i++ {
		sb.WriteString(keyPairs[i].key)
		if i != len(keyPairs)-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString(") VALUES (")
	for i := 0; i < len(keyPairs); i++ {
		sb.WriteString("$")
		sb.WriteString(fmt.Sprintf("%d", i+1))
		if i != len(keyPairs)-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func buildUpdate(tableName string, keyPairs, conditions []KeyPair) string {
	sb := strings.Builder{}
	sb.WriteString("UPDATE ")
	sb.WriteString(tableName)
	sb.WriteString(" SET ")
	for i := 0; i < len(keyPairs); i++ {
		sb.WriteString(keyPairs[i].key)
		sb.WriteString(" = $")
		sb.WriteString(fmt.Sprintf("%d", i+1))
		if i != len(keyPairs)-1 {
			sb.WriteString(",")
		}
	}
	sb.WriteString(" WHERE ")
	for i := 0; i < len(conditions); i++ {
		sb.WriteString(conditions[i].key)
		sb.WriteString(" = $")
		sb.WriteString(fmt.Sprintf("%d", i+len(keyPairs)+1))
		if i != len(conditions)-1 {
			sb.WriteString(" AND ")
		}
	}
	return sb.String()
}

func buildDelete(tableName string, conditions []KeyPair) string {
	sb := strings.Builder{}
	sb.WriteString("DELETE FROM ")
	sb.WriteString(tableName)
	sb.WriteString(" WHERE ")
	for i := 0; i < len(conditions); i++ {
		sb.WriteString(conditions[i].key)
		sb.WriteString(" = $")
		sb.WriteString(fmt.Sprintf("%d", i+1))
		if i != len(conditions)-1 {
			sb.WriteString(" AND ")
		}
	}
	return sb.String()
}

func (c *CockroachDB) QueryRow(tableName string, attributes []string, conditions []KeyPair) pgx.Row {
	row := c.conn.QueryRow(context.Background(), buildQuery(tableName, attributes, conditions))
	return row
}

func (c *CockroachDB) QueryRows(tableName string, attributes []string, conditions []KeyPair) pgx.Rows {
	rows, err := c.conn.Query(context.Background(), buildQuery(tableName, attributes, conditions))
	if err != nil {
		panic(err)
	}
	return rows
}

func (c *CockroachDB) Insert(tableName string, keyPairs []KeyPair) {
	values := make([]interface{}, len(keyPairs))
	for i := 0; i < len(keyPairs); i++ {
		values[i] = keyPairs[i].value
	}
	_, err := c.conn.Exec(context.Background(), buildInsert(tableName, keyPairs), values...)
	if err != nil {
		panic(err)
	}
}

func (c *CockroachDB) Update(tableName string, keyPairs, conditions []KeyPair) {
	_, err := c.conn.Exec(context.Background(), buildUpdate(tableName, keyPairs, conditions))
	if err != nil {
		panic(err)
	}
}

func (c *CockroachDB) Delete(tableName string, conditions []KeyPair) {
	_, err := c.conn.Exec(context.Background(), buildDelete(tableName, conditions))
	if err != nil {
		panic(err)
	}
}

func (c *CockroachDB) close() {
	c.conn.Close(context.Background())
}
