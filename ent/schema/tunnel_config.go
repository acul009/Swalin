package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Device holds the schema definition for the Device entity.
type TunnelConfig struct {
	ent.Schema
}

func (TunnelConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Bytes("config").NotEmpty(),
	}
}

func (TunnelConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		HostConfigMixin{},
	}
}
