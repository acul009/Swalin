package config

import (
	"context"
	"fmt"
	"rahnit-rmm/ent"
	"rahnit-rmm/util"

	"entgo.io/ent/dialect"
)

func openDB() (*ent.Client, error) {
	filepath := GetFilePath("db.sqlite")
	err := util.CreateParentDir(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %v", err)
	}

	client, err := ent.Open(dialect.SQLite, "file:"+filepath+"?mode=rwc&cache=shared&_fk=1")
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

var db *ent.Client

func InitDB() error {
	if db != nil {
		return fmt.Errorf("database already initialized")
	}
	client, err := openDB()
	if err != nil {
		return fmt.Errorf("failed opening connection to db: %v", err)
	}
	db = client
	return nil
}

func DB() *ent.Client {
	if db == nil {
		err := fmt.Errorf("database not initialized")
		panic(err)
	}
	return db
}
