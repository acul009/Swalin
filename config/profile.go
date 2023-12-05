package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rahn-it/svalin/db"
)

type Profile struct {
	subfolder string
	name      string
	dir       string
	db        *db.DB
	config    *Config
}

func OpenProfile(name string, subfolder string) (*Profile, error) {
	dir := filepath.Join(getConfigDir(), subfolder)

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	p := &Profile{
		subfolder: subfolder,
		name:      name,
		dir:       dir,
	}

	db, err := db.Open(p.getFilePath("bolt.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	p.db = db

	p.config = newConfig(p.Scope().Scope("config"))

	return p, nil
}

func ListProfiles(subfolder string) ([]string, error) {
	dir := filepath.Join(getConfigDir(), subfolder)

	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	p := &Profile{
		subfolder: subfolder,
		dir:       dir,
	}

	db, err := db.Open(p.getFilePath("bolt.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	list, err := db.ContextList()
	if err != nil {
		return nil, fmt.Errorf("failed to list context: %w", err)
	}

	stringList := make([]string, 0, len(list))
	for _, l := range list {
		stringList = append(stringList, string(l))
	}

	return stringList, nil
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
	return p.db.Context([]byte(p.name))
}

func (p *Profile) Config() *Config {
	return p.config
}
