package migrate

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var schema = Schema{
	Name: "test_schema",
	Changes: []Change{{
		Id: "create_customer_tables",
		Up: `CREATE TABLE customers(
			 	id SERIAL PRIMARY KEY NOT NULL
			 );`,
	}, {
		Id: "address_tables",
		Up: `CREATE TABLE addresses(
			 	id SERIAL PRIMARY KEY NOT NULL,
			 	customer_id INT REFERENCES customers(id) ON DELETE CASCADE
			 );`,
	}, {
		Id: "something_else",
		Up: `CREATE TABLE testing();`,
	}},
}

func TestSimpleMigration(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://localhost:5432/postgres?sslmode=disable&user=postgres")
	require.NoError(t, err)
	require.NoError(t, Sql(db).DestroyAndApply(context.Background(), &schema))
}
