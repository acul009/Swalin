package schema

import (
	"entgo.io/ent"
)

// Device holds the schema definition for the Device entity.
type TunnelConfig struct {
	ent.Schema
}

func (TunnelConfig) Fields() []ent.Field {
	return nil
}

func (TunnelConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		HostConfigMixin{},
	}
}
