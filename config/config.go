package config

import (
	"fmt"
	"strings"

	"github.com/rahn-it/svalin/db"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"go.etcd.io/bbolt"
)

type Config struct {
	scope    db.Scope
	flags    map[string]pflag.Value
	env      map[string]string
	defaults map[string]string
}

func newConfig(scope db.Scope) *Config {
	return &Config{
		scope:    scope,
		flags:    make(map[string]pflag.Value),
		env:      make(map[string]string),
		defaults: make(map[string]string),
	}
}

func (c *Config) BindFlags(flags *pflag.FlagSet) error {
	var err error
	flags.VisitAll(func(flag *pflag.Flag) {
		if err != nil {
			return
		}

		val := flag.Value
		if val == nil {
			err = fmt.Errorf("flag %s has no value", flag.Name)
			return
		}
		c.flags[strings.ToLower(flag.Name)] = val
	})

	return err
}

func (c *Config) Default(key string, value string) {
	cast := cast.ToString(value)
	c.defaults[strings.ToLower(key)] = cast
}

func (c *Config) Save(key string, value any) {
	cast := cast.ToString(value)
	err := c.scope.Update(func(b *bbolt.Bucket) error {
		return b.Put([]byte(key), []byte(cast))
	})

	if err != nil {
		panic(err)
	}
}

func (c *Config) String(key string) string {
	key = strings.ToLower(key)
	flag, ok := c.flags[key]
	if ok {
		return flag.String()
	}

	envVal, ok := c.env[key]
	if ok {
		return envVal
	}

	var dbVal []byte

	err := c.scope.View(func(b *bbolt.Bucket) error {
		dbVal = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		panic(err)
	}

	if dbVal != nil {
		return string(dbVal)
	}

	defaultVal, ok := c.defaults[key]
	if ok {
		return defaultVal
	}

	return ""
}
