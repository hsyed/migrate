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
	Changes: Changes{
		1: `
CREATE TABLE customers(
	id SERIAL PRIMARY KEY NOT NULL
);`,
		2: `
CREATE TABLE addresses(
	id SERIAL PRIMARY KEY NOT NULL,
	customer_id INT REFERENCES customers(id) ON DELETE CASCADE
);`,
		3: `
CREATE TABLE testing();
`,
	},
}

func TestSimpleMigration(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://localhost:5432/postgres?sslmode=disable&user=postgres")
	require.NoError(t, err)
	require.NoError(t, Sql(db).DestroyAndApply(context.Background(), &schema))
}
