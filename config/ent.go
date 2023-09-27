package config

import (
	"context"
	"fmt"
	"rahnit-rmm/ent"

	"entgo.io/ent/dialect"
)

func OpenDB() (*ent.Client, error) {
	client, err := ent.Open(dialect.SQLite, "file:"+GetFilePath("db.sqlite")+"?mode=rwc&cache=shared&_fk=1")
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to sqlite: %v", err)
	}

	ctx := context.Background()
	// Run the automatic migration tool to create all schema resources.
	if err := client.Schema.Create(ctx); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %v", err)
	}

	return client, nil
}
