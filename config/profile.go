package config

import (
	"fmt"
	"path/filepath"

	"github.com/rahn-it/svalin/db"
)

type Profile struct {
	name   string
	dir    string
	db     *db.DB
	config *Config
}

func OpenProfile(name string) (*Profile, error) {
	dir := filepath.Join(getConfigDir(), name)
	p := &Profile{
		name: name,
		dir:  dir,
	}

	db, err := db.Open(p.getFilePath("svalin.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	p.db = db

	p.config = newConfig(p.Scope().Scope("config"))

	return p, nil
}

func (p *Profile) Name() string {
	return p.name
}

func (p *Profile) getFilePath(filePath ...string) string {
	pathParts := make([]string, 1, len(filePath)+1)
	pathParts[0] = p.dir
	pathParts = append(pathParts, filePath...)
	fullPath := filepath.Join(pathParts...)
	return fullPath
}

func (p *Profile) DB() *db.DB {
	return p.db
}

func (p *Profile) Scope() db.Scope {
	return p.db.Scope([]byte(p.name))
}

func (p *Profile) Config() *Config {
	return p.config
}
