package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/mixin"
)

type HostConfigMixin struct {
	mixin.Schema
}

func (HostConfigMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("device", Device.Type).Unique().Required(),
	}
}
